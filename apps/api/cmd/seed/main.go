// Command seed upserts the demo example workflows (padel-court-booking,
// order-food, order-doctor) and their projects under a fixed demo tenant, so
// they can be resolved and executed end-to-end (PRD §40.1, epic #7).
//
// The seed is idempotent: re-running it does not duplicate rows (upsert by
// project slug / workflow slug). Demo scope is a dedicated tenant, so seeded
// data never pollutes other tenants (PRD §4).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
	"github.com/irosadie/open-state/go-shared/domain"
)

// Demo tenant/project identifiers. A fixed tenant UUID isolates seeded data.
const demoTenantID = "00000000-0000-0000-0000-000000000001"

// seedProject maps a demo project to its canonical workflow JSON file and the
// intent id that resolves to it (PRD §40.1).
type seedProject struct {
	slug              string
	name              string
	workflowFile      string
	intentID          string
	intentName        string
	intentDescription string
	examples          []string
}

var seeds = []seedProject{
	{
		slug: "padel", name: "Padel", workflowFile: "padel-booking.workflow.json", intentID: "BOOKING_PADEL",
		intentName: "Booking Lapangan Padel", intentDescription: "Memulai proses booking lapangan padel.",
		examples: []string{"saya mau order lapangan", "saya mau booking lapangan padel", "tolong carikan lapangan padel"},
	},
	{
		slug: "retail", name: "Retail", workflowFile: "order-food.workflow.json", intentID: "ORDER_FOOD",
		intentName: "Order Makanan", intentDescription: "Memulai proses pemesanan makanan.",
		examples: []string{"saya mau pesan makanan", "tolong order makanan", "saya mau pesan menu"},
	},
	{
		slug: "health", name: "Health", workflowFile: "order-doctor.workflow.json", intentID: "ORDER_DOCTOR",
		intentName: "Order Dokter", intentDescription: "Memulai proses pemesanan konsultasi dokter.",
		examples: []string{"saya mau konsultasi dokter", "tolong carikan dokter", "saya mau buat janji dokter"},
	},
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err.Error())
		return
	}
	slog.Info("seed complete")
}

func run() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	adapter := infradb.NewPostgresAdapter(pool)
	for _, s := range seeds {
		if err := seedOne(ctx, adapter.Projects(), adapter.Workflows(), adapter.Intents(), s); err != nil {
			return err
		}
	}
	if err := seedCapabilities(ctx, adapter.Capabilities()); err != nil {
		return err
	}
	return nil
}

// seedOne upserts a project, its workflow definition, and its intent mapping.
func seedOne(ctx context.Context, projects repositories.IProjectRepository, workflows repositories.IWorkflowRepository, intents repositories.IIntentRepository, s seedProject) error {
	proj, err := projects.FindBySlug(ctx, demoTenantID, s.slug)
	if err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("find project %s: %w", s.slug, err)
		}
		proj, err = projects.Create(ctx, demoTenantID, s.name, s.slug, entities.ProjectActive)
		if err != nil {
			return fmt.Errorf("create project %s: %w", s.slug, err)
		}
	}

	def, err := loadWorkflowDefinition(s.workflowFile, proj.ID)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("marshal workflow %s: %w", def.Slug, err)
	}

	wf, err := workflows.FindBySlug(ctx, demoTenantID, proj.ID, def.Slug)
	if err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("find workflow %s: %w", def.Slug, err)
		}
		desc := def.Description
		wf, err = workflows.Create(ctx, demoTenantID, proj.ID, def.Slug, def.Name, &desc, raw)
		if err != nil {
			return fmt.Errorf("create workflow %s: %w", def.Slug, err)
		}
	}

	versionNo := wf.CurrentVersion + 1
	if _, err := workflows.Publish(ctx, demoTenantID, proj.ID, wf.ID, versionNo, raw, entities.VersionStatusPublished, wf.Version); err != nil {
		return fmt.Errorf("publish workflow %s: %w", def.Slug, err)
	}
	if _, err := workflows.UpdateStatus(ctx, demoTenantID, proj.ID, wf.ID, entities.WorkflowPublished, wf.Version+1); err != nil {
		return fmt.Errorf("publish workflow status %s: %w", def.Slug, err)
	}
	if _, err := intents.Upsert(ctx, demoTenantID, proj.ID, wf.ID, s.intentID, s.intentName, s.intentDescription, s.examples); err != nil {
		return fmt.Errorf("upsert intent %s: %w", s.intentID, err)
	}

	slog.Info("seeded intent", "intent", s.intentID, "workflow", def.Slug, "project", s.slug)
	return nil
}

// loadWorkflowDefinition reads a canonical *.workflow.json, extracts the
// `workflow` envelope, and injects the projectId (engine format, PRD §161).
func loadWorkflowDefinition(file, projectID string) (*engine.WorkflowDefinition, error) {
	_, baseDir := repoRoot()
	raw, err := os.ReadFile(filepath.Join(baseDir, "docs", file))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	var envelope struct {
		Workflow engine.WorkflowDefinition `json:"workflow"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}
	def := envelope.Workflow
	def.ProjectID = projectID
	return &def, nil
}

// repoRoot returns the directory containing go.work (the repo root), by walking
// up from the current working directory.
func repoRoot() (string, string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", dir
		}
		dir = parent
	}
}

func isNotFound(err error) bool {
	var de *domain.DomainError
	return errors.As(err, &de) && de.Code == domain.ErrNotFound
}

// padelCapabilities defines the capabilities needed for the padel booking flow.
// These are seeded as INTERNAL/mock provider so they work with the JSONFileProvider.
var padelCapabilities = []struct {
	name        string
	description string
}{
	{"location.list", "Mengambil daftar cabang/lokasi padel yang tersedia"},
	{"padel.court.search", "Mencari lapangan padel berdasarkan lokasi dan tanggal"},
	{"padel.court.availability", "Mengecek ketersediaan slot waktu lapangan padel"},
	{"padel.court.book", "Membuat booking lapangan padel"},
	{"padel.payment.create", "Membuat payment link untuk booking"},
	{"padel.payment.verify", "Memverifikasi status pembayaran booking"},
	{"padel.notification.send", "Mengirim notifikasi konfirmasi booking"},
	{"food.menu.list", "Mengambil daftar menu restoran"},
	{"food.cart.add", "Menambahkan item ke keranjang pesanan"},
	{"food.cart.get", "Mengambil isi keranjang pesanan"},
	{"food.order.create", "Membuat pesanan makanan"},
	{"food.payment.create", "Membuat payment link untuk pesanan makanan"},
	{"food.payment.verify", "Memverifikasi pembayaran pesanan makanan"},
	{"food.order.track", "Melacak status pengiriman pesanan makanan"},
	{"doctor.search", "Mencari dokter berdasarkan spesialisasi dan lokasi"},
	{"doctor.schedule", "Mengambil jadwal praktik dokter"},
	{"doctor.book", "Membuat booking konsultasi dokter"},
	{"doctor.payment.create", "Membuat payment link untuk konsultasi dokter"},
	{"doctor.payment.verify", "Memverifikasi pembayaran konsultasi dokter"},
	{"doctor.notification.send", "Mengirim notifikasi konfirmasi booking dokter"},
	{"doctor.lookup", "Mencari dokter spesifik berdasarkan nama atau ID"},
	{"catalog.doctor_list", "Mengambil daftar dokter dari katalog berdasarkan spesialisasi"},
	{"schedule.check", "Mengecek jadwal dan ketersediaan slot dokter"},
	{"queue.check", "Mengecek ketersediaan antrian dokter untuk slot tertentu"},
	{"doctor.recommend", "Merekomendasikan dokter berdasarkan kebutuhan pasien"},
	{"booking.reserve", "Mereservasi slot booking dokter sementara"},
	{"booking.confirm", "Mengkonfirmasi booking dokter yang sudah direservasi"},
	{"booking.cancel", "Membatalkan booking dokter"},
	{"notification.send_confirmation", "Mengirim notifikasi konfirmasi booking dokter"},
}

// seedCapabilities upserts demo capabilities and binds them at tenant scope.
func seedCapabilities(ctx context.Context, caps repositories.ICapabilityRepository) error {
	for _, c := range padelCapabilities {
		desc := c.description
		providerServer, providerTool := demoProviderMapping(c.name)
		providerType := entities.ProviderTypeMCP
		cap, err := caps.FindByName(ctx, demoTenantID, c.name)
		if err != nil {
			if !isNotFound(err) {
				return fmt.Errorf("find capability %s: %w", c.name, err)
			}
			cap, err = caps.Create(ctx, demoTenantID, c.name, &desc,
				providerType, &providerServer, &providerTool,
				[]byte("{}"), []byte("{}"),
				1, nil,
			)
			if err != nil {
				return fmt.Errorf("create capability %s: %w", c.name, err)
			}
		} else if cap.ProviderType != providerType || cap.ProviderID.String != providerServer || cap.ProviderTool.String != providerTool {
			if cap, err = caps.Update(ctx, demoTenantID, cap.ID, &desc, providerType, &providerServer, &providerTool, cap.InputSchema, cap.OutputSchema, cap.Status, cap.Version, nil); err != nil {
				return fmt.Errorf("update capability %s: %w", c.name, err)
			}
		}
		// Bind at tenant scope (ALLOW) so it's accessible across all workflows.
		bindings, err := caps.ListBindingsByCapability(ctx, demoTenantID, cap.ID)
		if err != nil {
			return fmt.Errorf("list bindings %s: %w", c.name, err)
		}
		hasTenantBind := false
		for _, b := range bindings {
			if b.ScopeType == entities.BindingScopeTenant && b.ScopeID == demoTenantID {
				hasTenantBind = true
				break
			}
		}
		if !hasTenantBind {
			if _, err := caps.Bind(ctx, demoTenantID, cap.ID,
				entities.BindingScopeTenant, demoTenantID,
				entities.BindingPermissionAllow,
			); err != nil {
				return fmt.Errorf("bind capability %s: %w", c.name, err)
			}
		}
		slog.Info("seeded capability", "name", c.name, "id", cap.ID)
	}
	return nil
}

// demoProviderMapping keeps the demo workflow definitions free of URLs while
// mapping logical capabilities to the provider aliases exposed by the local
// provider mock scenarios. Production tenants configure their own aliases.
func demoProviderMapping(name string) (string, string) {
	switch {
	case strings.HasPrefix(name, "padel."), name == "location.list":
		return "padel-provider-mock", name
	case strings.HasPrefix(name, "food."):
		return "food-order-provider-mock", name
	default:
		return "doctor-provider-mock", name
	}
}
