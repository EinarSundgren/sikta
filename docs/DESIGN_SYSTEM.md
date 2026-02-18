# Design System

> Visual language, tokens, and component patterns. Fill in for each project.

---

## Overview

<!-- FILL IN: Brief description of the visual direction -->

**Design Direction:**  
<!-- e.g., "Minimal and functional", "Warm and approachable", "Technical and precise" -->

**Target Feeling:**  
<!-- e.g., "Professional but friendly", "Fast and efficient", "Calm and focused" -->

---

## Color Palette

<!-- FILL IN: Define your colors -->

### Primary Colors

| Name | Hex | Usage |
|------|-----|-------|
| `primary` | `#______` | Primary actions, links, focus states |
| `primary-hover` | `#______` | Hover state for primary |
| `primary-light` | `#______` | Backgrounds, subtle highlights |

### Neutral Colors

| Name | Hex | Usage |
|------|-----|-------|
| `gray-900` | `#______` | Primary text |
| `gray-700` | `#______` | Secondary text |
| `gray-500` | `#______` | Placeholder text, disabled |
| `gray-300` | `#______` | Borders |
| `gray-100` | `#______` | Backgrounds, cards |
| `white` | `#FFFFFF` | Page background |

### Semantic Colors

| Name | Hex | Usage |
|------|-----|-------|
| `success` | `#______` | Success states, confirmations |
| `warning` | `#______` | Warnings, caution |
| `error` | `#______` | Errors, destructive actions |
| `info` | `#______` | Informational messages |

### Tailwind Config

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#______',
          hover: '#______',
          light: '#______',
        },
        // Add other custom colors
      }
    }
  }
}
```

---

## Typography

<!-- FILL IN: Define your type system -->

### Font Families

| Role | Font | Fallback |
|------|------|----------|
| Body | `______` | `system-ui, sans-serif` |
| Headings | `______` | `system-ui, sans-serif` |
| Mono | `______` | `ui-monospace, monospace` |

### Type Scale

| Name | Size | Weight | Line Height | Usage |
|------|------|--------|-------------|-------|
| `h1` | `__px` / `__rem` | `___` | `___` | Page titles |
| `h2` | `__px` / `__rem` | `___` | `___` | Section titles |
| `h3` | `__px` / `__rem` | `___` | `___` | Subsections |
| `body` | `__px` / `__rem` | `___` | `___` | Default text |
| `small` | `__px` / `__rem` | `___` | `___` | Captions, meta |

### Tailwind Config

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['______', 'system-ui', 'sans-serif'],
        mono: ['______', 'ui-monospace', 'monospace'],
      },
      fontSize: {
        // Custom sizes if needed
      }
    }
  }
}
```

---

## Spacing

<!-- FILL IN: Define your spacing scale or use Tailwind defaults -->

Using Tailwind default spacing scale:

| Token | Value | Usage |
|-------|-------|-------|
| `1` | `0.25rem` (4px) | Tight spacing |
| `2` | `0.5rem` (8px) | Related elements |
| `4` | `1rem` (16px) | Standard gap |
| `6` | `1.5rem` (24px) | Section padding |
| `8` | `2rem` (32px) | Large gaps |
| `12` | `3rem` (48px) | Section margins |
| `16` | `4rem` (64px) | Page sections |

**Custom Spacing:**  
<!-- Add project-specific spacing tokens if needed -->

---

## Border Radius

<!-- FILL IN: Define your radius scale -->

| Name | Value | Usage |
|------|-------|-------|
| `sm` | `__px` | Subtle rounding |
| `DEFAULT` | `__px` | Buttons, inputs |
| `lg` | `__px` | Cards, modals |
| `full` | `9999px` | Circles, pills |

---

## Shadows

<!-- FILL IN: Define your shadow scale -->

| Name | Value | Usage |
|------|-------|-------|
| `sm` | `______` | Subtle depth |
| `DEFAULT` | `______` | Cards, dropdowns |
| `lg` | `______` | Modals, popovers |

---

## Component Patterns

<!-- FILL IN: Define reusable component patterns -->

### Buttons

```jsx
// Primary
<button className="px-4 py-2 bg-primary text-white rounded hover:bg-primary-hover">
  Primary Action
</button>

// Secondary
<button className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-100">
  Secondary
</button>

// Destructive
<button className="px-4 py-2 bg-error text-white rounded hover:opacity-90">
  Delete
</button>
```

### Inputs

```jsx
<input 
  type="text"
  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-primary"
  placeholder="Enter text..."
/>
```

### Cards

```jsx
<div className="p-6 bg-white border border-gray-200 rounded-lg shadow-sm">
  {/* Card content */}
</div>
```

### Alerts

```jsx
// Success
<div className="p-4 bg-green-50 border border-green-200 rounded text-green-800">
  Success message
</div>

// Error
<div className="p-4 bg-red-50 border border-red-200 rounded text-red-800">
  Error message
</div>
```

---

## Iconography

<!-- FILL IN: Icon library and usage conventions -->

**Icon Library:** <!-- e.g., Lucide, Heroicons, custom -->

**Sizes:**
| Name | Size | Usage |
|------|------|-------|
| `sm` | `16px` | Inline with text |
| `md` | `20px` | Buttons, list items |
| `lg` | `24px` | Standalone icons |

---

## Motion

<!-- FILL IN: Animation and transition guidelines -->

**Transition Duration:**
| Name | Duration | Usage |
|------|----------|-------|
| `fast` | `150ms` | Hovers, small changes |
| `normal` | `200ms` | Most transitions |
| `slow` | `300ms` | Large elements, modals |

**Easing:** `ease-in-out` (default)

---

## Responsive Breakpoints

Using Tailwind defaults:

| Prefix | Min Width | Typical Usage |
|--------|-----------|---------------|
| `sm` | `640px` | Large phones |
| `md` | `768px` | Tablets |
| `lg` | `1024px` | Laptops |
| `xl` | `1280px` | Desktops |
| `2xl` | `1536px` | Large screens |

---

## Dark Mode

<!-- FILL IN: Dark mode strategy if applicable -->

**Strategy:** <!-- e.g., "Not supported", "System preference", "User toggle" -->

<!-- If supporting dark mode, define dark variants for all colors -->

---

## Design Decisions Log

<!-- FILL IN: Record design decisions and rationale -->

| Date | Decision | Rationale |
|------|----------|-----------|
| — | Template initialized | Starting point |
