# CFC-TK Copilot Instructions

## Project Overview
CFC-TK is a SvelteKit 5 web application with a terminal emulator (xterm.js) and team management dashboard. Built with Vite, Tailwind CSS v4, TypeScript, and a custom UI component library (bits-ui based).

## Architecture

### Tech Stack
- **Framework**: SvelteKit 2.49+ (static adapter, Svelte 5.45+)
- **Styling**: Tailwind CSS 4.1 + Vite plugin, Tailwind Variants for component styles
- **UI Library**: Bits-UI 2.15.4 - headless components with Svelte bindings
- **Terminal**: xterm.js 6.0 with fit addon for responsive terminal sizing
- **Build**: Vite 7.2 with SvelteKit plugin

### Folder Structure
- `src/routes/` - SvelteKit pages (`+layout.svelte`, `+page.svelte`)
- `src/lib/components/ui/` - Reusable UI components (button, input, select, sheet, avatar, tooltip, label, separator)
- `src/lib/components/app/` - App-specific components (MainView with xterm, Sidebar, Topbar)
- `src/lib/utils.ts` - Shared utilities: `cn()` (clsx + tailwind-merge), type helpers
- Static files in `static/`

### Component Patterns

**UI Components** (`src/lib/components/ui/*/`):
- Each component has an `index.ts` barrel export
- Built with Bits-UI headless components + Tailwind styling
- Use `tv()` from `tailwind-variants` for variant management (e.g., button sizes, variants)
- Example: Button uses `buttonVariants` with variants like `default`, `destructive`, `outline`, `ghost`, `link`
- Compose complex components from subparts (e.g., `select` = Root + Group + Item + Content + Trigger)

**App Components**:
- `MainView.svelte` - Terminal emulator with dynamic xterm.js setup on mount
- `Sidebar.svelte`, `Topbar.svelte` - Navigation/layout
- Main layout: sidebar + topbar + content area structure

**Terminal Integration**:
- Lazily imported on mount (SSR safety via `browser` check)
- `FitAddon` observes container resize to fit terminal
- Basic local echo for user input (Backspace, Enter handling)
- Terminal writeln for text output

## Developer Workflow

### Commands
```bash
npm run dev          # Start Vite dev server
npm run build        # Production build
npm run preview      # Preview built app locally
npm run check        # Run svelte-check type checking
npm run check:watch  # Watch mode type checking
```

### Styling Approach
1. Use Tailwind utility classes for layout and basic styling
2. Use `cn()` utility to merge class conflicts: `cn("px-2", someCondition && "px-4")`
3. For component variants, use `tailwind-variants` `tv()` function (not inline conditional classes)
4. Import icons from `lucide-svelte` or `@lucide/svelte`

### UI Component Usage
Import from barrel exports:
```svelte
import { Button } from "$lib/components/ui/button";
import * as Select from "$lib/components/ui/select";
import { Label } from "$lib/components/ui/label";

<Button size="lg" variant="outline">Click me</Button>
<Label for="input">Name</Label>
<Select.Root bind:value={selectedTeam}>
  <Select.Trigger>...</Select.Trigger>
  <Select.Content>
    <Select.Item value="team1">Team 1</Select.Item>
  </Select.Content>
</Select.Root>
```

## Key Conventions

- **Type Safety**: Strict TypeScript enabled; use explicit types for props and state
- **Reactivity**: Use Svelte 5 `$state()`, `$derived()` (not two-way binding where unnecessary)
- **Aliases**: `$lib` points to `src/lib/`, `$app` for SvelteKit environment
- **SSR Safety**: Always check `browser` flag before using browser APIs (xterm, ResizeObserver, etc.)
- **CSS Variables**: Tailwind uses CSS custom properties for theme colors (primary, secondary, destructive, etc.)
- **Static Export**: App uses static adapter; avoid dynamic routes or server-side rendering assumptions

## Integration Points

- **Bits-UI**: Provides accessible, unstyled components; layer Tailwind + Bits-UI styling
- **xterm.js**: Terminal emulator loaded dynamically; hook ResizeObserver to FitAddon for responsiveness
- **Tailwind Variants**: Use `tv()` for complex component style combinations instead of nested conditionals
- **Mode Watcher**: Dark mode detection via `mode-watcher` plugin; applied via `data-sveltekit-preload-data`

## Code Examples

**Sheet (Modal) with Form**:
```svelte
<Sheet.Root bind:open={isOpen}>
  <Sheet.Content side="right">
    <Sheet.Header>
      <Sheet.Title>Add Team</Sheet.Title>
      <Sheet.Description>Fill in details</Sheet.Description>
    </Sheet.Header>
    <div class="grid gap-6">
      <Label for="name">Name</Label>
      <Input id="name" bind:value={teamName} />
    </div>
    <Sheet.Footer>
      <Button type="submit">Save</Button>
      <Sheet.Close class={buttonVariants({ variant: "outline" })}>Cancel</Sheet.Close>
    </Sheet.Footer>
  </Sheet.Content>
</Sheet.Root>
```

**xterm.js Setup**:
```svelte
import { onMount, browser } from "svelte";
onMount(async () => {
  if (!browser) return;
  const { Terminal } = await import("xterm");
  const { FitAddon } = await import("xterm-addon-fit");
  const term = new Terminal({ cursorBlink: true, fontSize: 13 });
  const fit = new FitAddon();
  term.loadAddon(fit);
  term.open(containerElement);
  fit.fit();
  new ResizeObserver(() => fit.fit()).observe(containerElement);
});
```

## Common Tasks

- **Add UI Component**: Copy structure from existing (e.g., button), update `index.ts` exports and variant styles
- **Add Page**: Create `src/routes/+page.svelte` or subdirectory with `+page.svelte`
- **Update Styles**: Edit Tailwind classes or `tv()` variants in component files
- **Debug Terminal**: Check xterm logs in browser DevTools; verify FitAddon observer is active
