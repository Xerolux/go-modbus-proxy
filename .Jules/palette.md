## 2026-01-10 - Keyboard Accessibility on Dashboard
**Learning:** Icon-only `div` "buttons" are a common pattern that completely excludes keyboard users and screen readers.
**Action:** Always replace interactive `div`s with semantic `<button>` elements, ensuring they have `aria-label` (if icon-only) and visible focus states (e.g., `focus:ring`).
