# Markdown Renderer mit Plugin-System

## Task

Implement `markdown-renderer.js` — the integration layer that connects two fully implemented modules:

- `renderer.js` — Markdown block parser and inline transformer
- `plugin-system.js` — Plugin registry for custom syntax directives

**Do not modify `renderer.js`, `plugin-system.js`, or the test files.**

### Available API from `renderer.js`

```javascript
const { parseBlocks, applyInline, serializeBlocks } = require('./renderer');
```

#### `parseBlocks(markdown)`

Parses a Markdown string into an array of block objects. Returns `[]` for empty or whitespace-only input.

Block object shapes:

| `type`       | Additional fields                              |
|--------------|------------------------------------------------|
| `'heading1'` | `content: string`                              |
| `'heading2'` | `content: string`                              |
| `'heading3'` | `content: string`                              |
| `'paragraph'`| `content: string`                              |
| `'blockquote'`| `content: string`                             |
| `'hr'`       | *(no additional fields)*                       |
| `'code'`     | `content: string`, `language: string` (may be `''`) |
| `'ul'`       | `items: string[]`                              |
| `'ol'`       | `items: string[]`                              |

Consecutive list lines of the same type are collected into one block. Empty lines between blocks are ignored.

#### `applyInline(text)`

Applies inline transformations to a string:

| Markdown       | HTML                           |
|----------------|--------------------------------|
| `**text**`     | `<strong>text</strong>`        |
| `*text*`       | `<em>text</em>`                |
| `` `code` ``   | `<code>code</code>`            |
| `[text](url)`  | `<a href="url">text</a>`       |

Inline code (backtick) suppresses bold and italic parsing of its content.

#### `serializeBlocks(blocks)`

Serializes a block array to an HTML string **without** applying inline transformations. Use this only if you want raw serialization; for the full rendering pipeline, apply `applyInline()` yourself before or after serialization as appropriate.

---

### Available API from `plugin-system.js`

```javascript
const { createRegistry } = require('./plugin-system');
```

#### `createRegistry()`

Creates a new plugin registry. Returns an object with:

- `register(name, handler)` — Registers a plugin. Duplicate names overwrite the previous registration.
- `resolve(text)` — Replaces all plugin directives in the text with plugin output. Unknown directives are left unchanged.
- `list()` — Returns an array of all registered plugin names.

**Plugin directive syntax:**

```
::pluginname[content]              → handler('content', [])
::pluginname[attribute][content]   → handler('content', ['attribute'])
```

---

### Functions to Implement in `markdown-renderer.js`

#### `renderMarkdown(markdown, options = {})`

Renders a Markdown string to HTML using `parseBlocks()` and `applyInline()`.

- Call `parseBlocks(markdown)` to obtain the block array
- For each block, produce the corresponding HTML:

  | Block type   | HTML output                                              |
  |--------------|----------------------------------------------------------|
  | `heading1`   | `<h1>INLINE</h1>`                                        |
  | `heading2`   | `<h2>INLINE</h2>`                                        |
  | `heading3`   | `<h3>INLINE</h3>`                                        |
  | `paragraph`  | `<p>INLINE</p>`                                          |
  | `blockquote` | `<blockquote>INLINE</blockquote>`                        |
  | `hr`         | `<hr>`                                                   |
  | `code` (no language) | `<pre><code>CONTENT</code></pre>`               |
  | `code` (with language `js`) | `<pre><code class="language-js">CONTENT</code></pre>` |
  | `ul`         | `<ul><li>INLINE</li><li>INLINE</li>...</ul>`             |
  | `ol`         | `<ol><li>INLINE</li><li>INLINE</li>...</ol>`             |

  Where `INLINE` means the content has been passed through `applyInline()`, and `CONTENT` means the content is left **as-is** (no inline transformation for code blocks).

- Concatenate all HTML strings and return the result
- Return `''` for empty or whitespace-only input

#### `createRenderer(options = {})`

Returns a `RendererInstance` object:

```javascript
{
  use(plugin),   // chainable
  render(markdown)
}
```

- Internally call `createRegistry()` once to create a registry for this instance
- `use(plugin)`:
  - Validate the plugin: throw `TypeError` if `plugin` is not an object, or if `plugin.name` is not a string, or if `plugin.handler` is not a function
  - Call `registry.register(plugin.name, plugin.handler)`
  - Return `this` (to allow method chaining)
- `render(markdown)`:
  - Call `renderMarkdown(markdown)` to produce HTML
  - Call `registry.resolve(html)` to replace plugin directives
  - Return the result

#### `createPlugin(name, handler)`

Creates and returns a plugin descriptor object `{ name, handler }`.

- Throw `TypeError` if `handler` is not a function
- Return `{ name, handler }`

## Context

In real-world software development, building new features from scratch is less common than **connecting existing components**. Libraries, SDKs, and internal modules expose stable APIs — the engineering challenge is reading unfamiliar code, understanding its contracts, and wiring the parts together correctly.

This challenge replicates that scenario. `renderer.js` and `plugin-system.js` are production-ready modules with documented APIs. Your task is to implement the integration layer that:

1. Understands and uses the block model from `parseBlocks()`
2. Applies inline transformations correctly (block content vs. code content)
3. Builds a fluent plugin interface on top of the registry
4. Respects API contracts: `TypeError` on invalid input, `''` on empty input, method chaining via `return this`

Skills exercised:

- Reading and integrating third-party APIs
- Multi-step rendering pipelines (parse → transform → serialize → resolve)
- Fluent interface / method chaining patterns
- Higher-order functions (plugin handler delegation)
- Edge case handling: empty input, code block protection, unknown directives

## Dependencies

- Node.js 18 or newer
- No external runtime npm packages — use only the provided `renderer.js` and `plugin-system.js` modules
- Jest 29 is provided for testing (via `devDependencies` in `package.json`)

Local commands:

```bash
npm install
npm test
```

For watch mode:

```bash
npm run test:watch
```

## Constraints

- **Do not modify `renderer.js` or `plugin-system.js`** — they are fully implemented and correct
- **Do not modify the test files**
- Do not use any external npm packages in `markdown-renderer.js`
- `use()` must return `this` (the renderer instance) to enable method chaining
- `renderMarkdown()` must not know about or use the plugin system
- `applyInline()` must **not** be called on the content of `code` blocks
- Plugin directives are resolved **after** `renderMarkdown()` runs (via `registry.resolve()`)
- `createPlugin()` must throw `TypeError` when `handler` is not a function
- `use()` must throw `TypeError` when the plugin object is missing `name` or `handler`

## Edge Cases

- Empty string input to `renderMarkdown` or `render` → return `''`
- Whitespace-only input → return `''`
- Heading with inline formatting: `# Hello **world**` → `<h1>Hello <strong>world</strong></h1>`
- Code block content must not be transformed: `**bold**` inside a code block stays as `**bold**`
- Plugin directive inside a code block must not be resolved (the `resolve()` call runs on the full output after rendering, but code block content is already wrapped in `<pre><code>...</code></pre>` — the directive syntax `::name[...]` will not appear as an unresolved string in the final HTML if the code block was rendered correctly)
- Unknown plugin directive (no plugin registered for that name) → directive left unchanged, no error
- `createPlugin('x', 'not-a-function')` → `TypeError`
- `createPlugin('x', null)` → `TypeError`
- `renderer.use({})` → `TypeError` (missing `name` and `handler`)
- `renderer.use({ name: 'x' })` → `TypeError` (missing `handler`)
- `renderer.use(null)` → `TypeError`
- Registering a plugin with a duplicate name replaces the previous registration
- Multiple plugins chained with `.use().use().use()` — all must be active in a single `.render()` call
- `renderMarkdown()` called directly must not resolve plugin directives (no plugin system involved)
