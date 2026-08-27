---
name: web-seo-geo-friendly
description: Optimize public Next.js App Router pages for SEO and GEO (Generative Engine Optimization). Use when the user explicitly requests SEO/GEO optimization on public pages.
---

# Skill: Web SEO + GEO

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

Optimize public pages for discoverability by search engines and AI answer engines.

## Scope

Applies only to **public** pages in `apps/web/app/` that are not protected by auth.

## Workflow

1. Add `generateMetadata` or `metadata` in `page.tsx`
2. Add Open Graph and Twitter card meta
3. Add JSON-LD structured data matching the content type
4. Add the page to `sitemap.ts`
5. Ensure `robots.ts` does not block public pages

## Prohibitions

- **FORBIDDEN** to apply this skill to private or auth-protected pages.
- **FORBIDDEN** to add generic metadata unrelated to the page content.
- **FORBIDDEN** to add structured data that is invalid or irrelevant to the content.

## Pre-Completion Checklist

- [ ] `generateMetadata` present: title, description, canonical
- [ ] Open Graph meta: title, description, url, image
- [ ] Twitter card meta: card type, title, description
- [ ] JSON-LD structured data matches the page type
- [ ] Page listed in `sitemap.ts`
- [ ] `robots.ts` does not block the page
- [ ] H1 present and descriptive (one per page)
- [ ] Meta description 150-160 characters
- [ ] Every file ends with a newline (EOF)
