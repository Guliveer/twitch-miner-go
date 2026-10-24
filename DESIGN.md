# Modern Dark

## Identity

Use a near-black page, quiet grey reading text, and bright white statements. Broad pools of white light reveal a small number of opaque graphite objects. Keep the typography regular and precise. A tiny mint signal can identify a live or selected state; it is not a general accent color.

## Layout

Work on a twelve-column grid with a shared maximum width and generous gutters. Let a large opening statement occupy most of the left side while a narrower product object sits to the right, partially sharing its space. Keep the object clear of readable copy. A ruled row of facts can anchor the bottom of this opening when the brief supplies useful facts.

Carry the same column edges through the page. Available arrangements include an unequal open feature grid, one large editorial image beside two smaller image-and-text rows, and alternating pairs of explanatory text and framed product views. Give section headings more space than the compact information beneath them. A large closing statement can share its field with a cropped graphite object and a softer pool of white light. Use a ruled footer with clearly grouped links.

Use only the sections and interface views the customer needs. The decorative chart, bars, and nodes are examples of a product object, not required application features or data.

## Typography

Use Geist for display and reading, and Geist Mono for labels, dates, numerical values, and short notes. The aliases "Modern Dark Grotesk" and "Modern Dark Mono" refer to those faces. Keep large type at regular weight with tight tracking: the opening display scales from 44px to 104px with leading near one. Section titles are smaller, while body copy stays relaxed. Distinguish leading phrases with bright text and supporting phrases with a readable grey; do not fade essential information into the background.

## Color and surfaces

The ground token is the darkest field. Panel and hairline roles separate opaque surfaces without bright outlines. Use bright for titles and primary actions, read for paragraphs, and meta for small supporting text. Restrict live to small status marks with an accompanying label or state.

White light may fall across the hero and closing object through soft radial gradients. Product objects can have subtle grey material gradients. Keep ordinary reading areas flat. Use the small panel radius for imagery and frames, and the pill radius for buttons; these serve different roles.

## Components

Use compact pill buttons with a white filled primary action and a dark outlined secondary action. Hover changes fill or border; focus remains a visible offset outline. Frame product views as opaque graphite panels. Keep functional interface content truthful and legible.

When comparisons are needed, use aligned columns with a quiet filled featured column. Show selection separately with a labeled control and a small signal. Disclosures use a numbered gutter, question, and plus/minus within ruled rows. These are available patterns, not required sections.

## Images

Use grayscale editorial photography with restrained contrast and slightly darkened image fields. A lead image can use a tall 4:5 crop, with square supporting images beside summaries. Small round grayscale portraits can identify real people. Preserve recognizable subjects and reserve image dimensions. Product screenshots should communicate the customer's actual product rather than reproduce abstract sample charts.

## Motion

Use the duration and easing tokens for brief control color changes. The white lighting is static. Do not add pointer-following light, staged entrances, floating panels, or animated decorative charts.

## Responsive behavior

Respond to the containing block. At 1180px, reduce gutters and give the opening copy the full grid width. Around 900px, simplify multi-column arrangements, use two-column facts and short-entry grids, and stack comparison options. At 560px, use the narrow gutter, single-column entries and quotations, and images above their text with wider crops. Scale decorative objects without covering copy. Keep the complete footer and all actions available.

## Design limits

Keep large areas neutral and opaque. Avoid colored glows, translucent glass, heavy display weights, and decorative motion. Do not invent metrics, testimonials, application states, or extra sections to fill the available patterns.

## Authoritative tokens

| CSS variable | Value |
| --- | --- |
| --dp-color-ground | #050505 |
| --dp-color-panel | #0e0e0e |
| --dp-color-hairline | #222222 |
| --dp-color-hairline-2 | #303030 |
| --dp-color-bright | #f2f2f2 |
| --dp-color-read | #a6a9ad |
| --dp-color-meta | #8a8d91 |
| --dp-color-live | #3fd6c6 |
| --dp-color-inverse-ink | #0a0a0a |
| --dp-color-dim | #96999d |
| --dp-typography-face | "Modern Dark Grotesk", "Helvetica Neue", Arial, sans-serif |
| --dp-typography-mono | "Modern Dark Mono", "SFMono-Regular", Menlo, monospace |
| --dp-spacing-gutter | 48px |
| --dp-spacing-gap | 24px |
| --dp-spacing-max | 1360px |
| --dp-radius-panel | 6px |
| --dp-radius-pill | 999px |
| --dp-border-hairline | 1px |
| --dp-motion-duration | 120ms |
| --dp-motion-ease | ease-out |