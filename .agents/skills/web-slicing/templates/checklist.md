# Checklist: Web Slicing

- [ ] Read `.agents/settings.json`
- [ ] Read `.agents/guides/ARCHITECTURE.md` (apps/web section)
- [ ] Read `references/context.md`
- [ ] Read `.agents/guides/web-page.md` before creating files in `app/`
- [ ] Read `.agents/guides/web-component.md` before creating files in `components/`
- [ ] Review examples at `.agents/examples/web-slicing/nextjs-app-router/`
- [ ] `page.tsx` is a thin Suspense wrapper — no logic
- [ ] `*-page-content.tsx` is beside `page.tsx` and orchestrates the route
- [ ] Components private to the route are in `_components/`; shared components are in `apps/web/components/`
- [ ] No `axios`/`fetch` in components
- [ ] Search: `useQueryParam` + `debounce` utility + `useMemo`
- [ ] Delete: SweetAlert `preConfirm` + `new Promise` + `.then()/.catch()`
- [ ] Toast: `react-hot-toast` (`toast.success`, `toast.error`)
- [ ] No `any`
- [ ] Every file ends with a newline (EOF)
