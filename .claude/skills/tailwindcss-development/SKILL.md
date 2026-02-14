---
name: tailwindcss-development
description: >-
  Styles applications using Tailwind CSS v4 utilities. Activates when adding styles, restyling components,
  working with gradients, spacing, layout, flex, grid, responsive design, dark mode, colors,
  typography, or borders; or when the user mentions CSS, styling, classes, Tailwind, restyle,
  hero section, cards, buttons, or any visual/UI changes.
---

# Tailwind CSS Development

## When to Apply

Activate this skill when:

- Adding styles to components or pages
- Working with responsive design
- Implementing dark mode
- Extracting repeated patterns into components
- Debugging spacing or layout issues

## Project Setup

This project uses **Tailwind CSS v4** with a build step via `@tailwindcss/cli`. No CDN.

- **Source CSS**: `web/static/css/input.css` — contains `@import "tailwindcss"`, `@theme`, `@variant`, and CSS variables
- **Compiled CSS**: `web/static/css/style.css` — generated output (gitignored)
- **Build**: `npm run build:css` (minified) / `npm run dev:css` (watch mode)
- **Config**: CSS-first via `@theme` in `input.css` (no `tailwind.config.js`)
- **Dark mode**: Class-based via `@variant dark (&:where(.dark, .dark *));`
- **Templating**: Go `html/template` with `{{define}}`, `{{template}}`, `{{block}}` directives
- **Templates**: `web/templates/layouts/base.html` (base layout), `web/templates/*.html` (pages)
- **Static serving**: Gin serves `/static` from `./static` directory
- **PWA**: `web/static/manifest.json` (app manifest) and `web/static/js/sw.js` (service worker) are linked from `base.html`. Theme color in the manifest (`#b8922f`) should stay in sync with `primary-500`.

## Basic Usage

- Use Tailwind CSS v4 utility classes to style HTML. Check and follow existing conventions in the project before introducing new patterns.
- Offer to extract repeated patterns into Go template blocks or partials that match the project's conventions.
- Consider class placement, order, priority, and defaults. Remove redundant classes, add classes to parent or child elements carefully to reduce repetition, and group elements logically.
- After modifying templates, run `npm run build:css` to regenerate the compiled CSS.

## Tailwind CSS v4 Specifics

- Always use Tailwind CSS v4 and avoid deprecated utilities.
- `corePlugins` and `tailwind.config.js` are not supported in Tailwind v4.
- `@apply` is available (build-based setup supports it).

### CSS-First Configuration

In Tailwind v4, configuration is CSS-first using the `@theme` directive in `static/css/input.css`:

<code-snippet name="CSS-First Config" lang="css">
@theme {
  --color-brand: oklch(0.72 0.11 178);
}
</code-snippet>

### Import Syntax

In Tailwind v4, import Tailwind with a regular CSS `@import` statement instead of the `@tailwind` directives used in v3:

<code-snippet name="v4 Import Syntax" lang="diff">
- @tailwind base;
- @tailwind components;
- @tailwind utilities;
+ @import "tailwindcss";
</code-snippet>

### Replaced Utilities

Tailwind v4 removed deprecated utilities. Use the replacements shown below. Opacity values remain numeric.

| Deprecated | Replacement |
|------------|-------------|
| bg-opacity-* | bg-black/* |
| text-opacity-* | text-black/* |
| border-opacity-* | border-black/* |
| divide-opacity-* | divide-black/* |
| ring-opacity-* | ring-black/* |
| placeholder-opacity-* | placeholder-black/* |
| flex-shrink-* | shrink-* |
| flex-grow-* | grow-* |
| overflow-ellipsis | text-ellipsis |
| decoration-slice | box-decoration-slice |
| decoration-clone | box-decoration-clone |

## Theme System

This project uses a **CSS variable-based OKLCH color theme** with 5 semantic color scales, each with shades 50–950:

| Scale | Purpose |
|-------|---------|
| `text` | Text/foreground colors |
| `background` | Page and surface backgrounds |
| `primary` | Primary brand/action color |
| `secondary` | Supporting/complementary color |
| `accent` | Highlights and calls to attention |

### How It Works

1. CSS custom properties (e.g. `--text-50`, `--background-900`) hold OKLCH color parameters
2. `:root` defines light mode values, `.dark` defines dark mode values (scales are inverted)
3. `@theme` registers these as Tailwind colors via `--color-*` tokens referencing the variables
4. Tailwind generates utilities like `bg-background-50`, `text-text-900`, `border-primary-500`

### Usage Examples

<code-snippet name="Theme Colors" lang="html">
<!-- Page background — adapts to light/dark automatically -->
<div class="bg-background-50">

<!-- Heading text — adapts automatically -->
<h1 class="text-text-900">

<!-- Button with primary color -->
<button class="bg-primary-500 text-white">

<!-- Accent border -->
<div class="border border-accent-400">
</code-snippet>

## Dark Mode

Dark mode is handled **automatically** by the CSS variable system. The `:root` and `.dark` selectors swap the OKLCH values for each color scale, so a single class handles both modes:

<code-snippet name="Dark Mode — Correct" lang="html">
<!-- CORRECT: CSS variables handle the light/dark swap -->
<div class="bg-background-50 text-text-900">
</code-snippet>

<code-snippet name="Dark Mode — Unnecessary" lang="html">
<!-- UNNECESSARY: dark: prefix is redundant for themed colors -->
<div class="bg-background-50 dark:bg-background-950 text-text-900 dark:text-text-50">
</code-snippet>

Only use the `dark:` variant when you need a fundamentally different style in dark mode that the CSS variable swap doesn't cover (e.g. a different shadow, or a non-themed color like `bg-white dark:bg-background-300`).

## Spacing

Use `gap` utilities instead of margins for spacing between siblings:

<code-snippet name="Gap Utilities" lang="html">
<div class="flex gap-8">
    <div>Item 1</div>
    <div>Item 2</div>
</div>
</code-snippet>

## Common Patterns

### Flexbox Layout

<code-snippet name="Flexbox Layout" lang="html">
<div class="flex items-center justify-between gap-4">
    <div>Left content</div>
    <div>Right content</div>
</div>
</code-snippet>

### Grid Layout

<code-snippet name="Grid Layout" lang="html">
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    <div>Card 1</div>
    <div>Card 2</div>
    <div>Card 3</div>
</div>
</code-snippet>

## Common Pitfalls

- Using deprecated v3 utilities (bg-opacity-*, flex-shrink-*, etc.)
- Using `@tailwind` directives instead of `@import "tailwindcss"`
- Trying to use `tailwind.config.js` instead of CSS `@theme` directive
- Using margins for spacing between siblings instead of gap utilities
- Adding unnecessary `dark:` prefixes for themed colors (the CSS variables already handle it)
- Forgetting to run `npm run build:css` after modifying templates or `input.css`
