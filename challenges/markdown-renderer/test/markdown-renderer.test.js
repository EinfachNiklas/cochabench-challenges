'use strict';

const {
  renderMarkdown,
  createRenderer,
  createPlugin
} = require('../markdown-renderer');

// ---------------------------------------------------------------------------
// Helper plugins for tests
// ---------------------------------------------------------------------------

const warningPlugin = createPlugin('warning', (content) =>
  `<div class="warning">${content}</div>`
);

const badgePlugin = createPlugin('badge', (content, attrs) => {
  const type = attrs && attrs[0] ? attrs[0] : '';
  return `<span class="badge badge-${type}">${content}</span>`;
});

const uppercasePlugin = createPlugin('upper', (content) =>
  `<span class="upper">${content.toUpperCase()}</span>`
);

// ---------------------------------------------------------------------------
// renderMarkdown – Headings
// ---------------------------------------------------------------------------

describe('renderMarkdown – Headings', () => {
  test('renders H1', () => {
    expect(renderMarkdown('# Hello')).toBe('<h1>Hello</h1>');
  });

  test('renders H2', () => {
    expect(renderMarkdown('## Hello')).toBe('<h2>Hello</h2>');
  });

  test('renders H3', () => {
    expect(renderMarkdown('### Hello')).toBe('<h3>Hello</h3>');
  });

  test('renders H1 with inline bold', () => {
    expect(renderMarkdown('# Hello **World**')).toBe('<h1>Hello <strong>World</strong></h1>');
  });

  test('renders H2 with inline italic', () => {
    expect(renderMarkdown('## Hello *World*')).toBe('<h2>Hello <em>World</em></h2>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Paragraphs
// ---------------------------------------------------------------------------

describe('renderMarkdown – Paragraphs', () => {
  test('renders simple paragraph', () => {
    expect(renderMarkdown('Hello World')).toBe('<p>Hello World</p>');
  });

  test('empty string returns empty string', () => {
    expect(renderMarkdown('')).toBe('');
  });

  test('whitespace only returns empty string', () => {
    expect(renderMarkdown('   ')).toBe('');
  });

  test('renders multiple paragraphs', () => {
    const input = 'First paragraph\n\nSecond paragraph';
    const output = renderMarkdown(input);
    expect(output).toContain('<p>First paragraph</p>');
    expect(output).toContain('<p>Second paragraph</p>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Inline elements
// ---------------------------------------------------------------------------

describe('renderMarkdown – Inline elements', () => {
  test('renders bold', () => {
    expect(renderMarkdown('**bold**')).toContain('<strong>bold</strong>');
  });

  test('renders italic', () => {
    expect(renderMarkdown('*italic*')).toContain('<em>italic</em>');
  });

  test('renders inline code', () => {
    expect(renderMarkdown('`code`')).toContain('<code>code</code>');
  });

  test('renders link', () => {
    expect(renderMarkdown('[Click](https://example.com)')).toContain(
      '<a href="https://example.com">Click</a>'
    );
  });

  test('combined inline elements in paragraph', () => {
    const output = renderMarkdown('Text with **bold** and *italic* and `code`');
    expect(output).toContain('<strong>bold</strong>');
    expect(output).toContain('<em>italic</em>');
    expect(output).toContain('<code>code</code>');
  });

  test('inline code suppresses bold parsing', () => {
    const output = renderMarkdown('`**no bold**`');
    expect(output).toContain('<code>**no bold**</code>');
    expect(output).not.toContain('<strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Code blocks
// ---------------------------------------------------------------------------

describe('renderMarkdown – Code blocks', () => {
  test('renders code block without language', () => {
    const input = '```\nconsole.log("hi");\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('<pre>');
    expect(output).toContain('<code>');
    expect(output).toContain('console.log("hi");');
  });

  test('renders code block with language specified', () => {
    const input = '```javascript\nconst x = 1;\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('class="language-javascript"');
    expect(output).toContain('const x = 1;');
  });

  test('no inline rendering inside code blocks', () => {
    const input = '```\n**not bold**\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('**not bold**');
    expect(output).not.toContain('<strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Lists
// ---------------------------------------------------------------------------

describe('renderMarkdown – Lists', () => {
  test('renders unordered list', () => {
    const input = '- One\n- Two\n- Three';
    const output = renderMarkdown(input);
    expect(output).toContain('<ul>');
    expect(output).toContain('<li>One</li>');
    expect(output).toContain('<li>Two</li>');
    expect(output).toContain('<li>Three</li>');
    expect(output).toContain('</ul>');
  });

  test('renders ordered list', () => {
    const input = '1. Alpha\n2. Beta\n3. Gamma';
    const output = renderMarkdown(input);
    expect(output).toContain('<ol>');
    expect(output).toContain('<li>Alpha</li>');
    expect(output).toContain('<li>Beta</li>');
    expect(output).toContain('<li>Gamma</li>');
    expect(output).toContain('</ol>');
  });

  test('consecutive list items produce a single list', () => {
    const input = '- A\n- B';
    const output = renderMarkdown(input);
    const ulCount = (output.match(/<ul>/g) || []).length;
    expect(ulCount).toBe(1);
  });

  test('list items support inline elements', () => {
    const input = '- **bold** item';
    const output = renderMarkdown(input);
    expect(output).toContain('<strong>bold</strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Blockquotes
// ---------------------------------------------------------------------------

describe('renderMarkdown – Blockquotes', () => {
  test('renders simple blockquote', () => {
    expect(renderMarkdown('> Quote')).toContain('<blockquote>Quote</blockquote>');
  });

  test('renders blockquote with inline formatting', () => {
    const output = renderMarkdown('> **important**');
    expect(output).toContain('<blockquote>');
    expect(output).toContain('<strong>important</strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Horizontal rule
// ---------------------------------------------------------------------------

describe('renderMarkdown – Horizontal rule', () => {
  test('renders horizontal rule', () => {
    expect(renderMarkdown('---')).toContain('<hr>');
  });

  test('--- in paragraph context does not create hr', () => {
    const output = renderMarkdown('Text --- Text');
    expect(output).not.toContain('<hr>');
  });
});

// ---------------------------------------------------------------------------
// createPlugin
// ---------------------------------------------------------------------------

describe('createPlugin', () => {
  test('creates valid plugin object', () => {
    const plugin = createPlugin('test', (content) => content);
    expect(plugin).toHaveProperty('name', 'test');
    expect(plugin).toHaveProperty('handler');
    expect(typeof plugin.handler).toBe('function');
  });

  test('throws TypeError when handler is not a function', () => {
    expect(() => createPlugin('test', 'not-a-handler')).toThrow(TypeError);
  });

  test('throws TypeError when handler is null', () => {
    expect(() => createPlugin('test', null)).toThrow(TypeError);
  });

  test('handler is called correctly', () => {
    const plugin = createPlugin('greet', (content) => `Hello ${content}`);
    expect(plugin.handler('World', [])).toBe('Hello World');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – Basic rendering
// ---------------------------------------------------------------------------

describe('createRenderer – Basic rendering', () => {
  test('renderer without plugins behaves like renderMarkdown', () => {
    const renderer = createRenderer();
    expect(renderer.render('# Title')).toBe(renderMarkdown('# Title'));
  });

  test('renderer returns empty string for empty input', () => {
    const renderer = createRenderer();
    expect(renderer.render('')).toBe('');
  });

  test('renderer renders standard Markdown correctly', () => {
    const renderer = createRenderer();
    const output = renderer.render('**bold** and *italic*');
    expect(output).toContain('<strong>bold</strong>');
    expect(output).toContain('<em>italic</em>');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – Plugin registration
// ---------------------------------------------------------------------------

describe('createRenderer – Plugin registration', () => {
  test('use() returns the instance (chainable)', () => {
    const renderer = createRenderer();
    const result = renderer.use(warningPlugin);
    expect(result).toBe(renderer);
  });

  test('use() with null throws TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use(null)).toThrow(TypeError);
  });

  test('use() with empty object throws TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({})).toThrow(TypeError);
  });

  test('use() with missing handler throws TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({ name: 'x' })).toThrow(TypeError);
  });

  test('use() with missing name throws TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({ handler: () => {} })).toThrow(TypeError);
  });

  test('duplicate plugin name overwrites previous registration', () => {
    const renderer = createRenderer();
    const plugin1 = createPlugin('tag', () => '<span>v1</span>');
    const plugin2 = createPlugin('tag', () => '<span>v2</span>');
    renderer.use(plugin1).use(plugin2);
    const output = renderer.render('::tag[x]');
    expect(output).toContain('<span>v2</span>');
    expect(output).not.toContain('<span>v1</span>');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – warningPlugin
// ---------------------------------------------------------------------------

describe('createRenderer – warningPlugin', () => {
  test('renders ::warning[...] correctly', () => {
    const renderer = createRenderer().use(warningPlugin);
    expect(renderer.render('::warning[Attention!]')).toContain(
      '<div class="warning">Attention!</div>'
    );
  });

  test('::warning with empty content', () => {
    const renderer = createRenderer().use(warningPlugin);
    expect(renderer.render('::warning[]')).toContain('<div class="warning"></div>');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – badgePlugin
// ---------------------------------------------------------------------------

describe('createRenderer – badgePlugin', () => {
  test('renders ::badge[success][Done]', () => {
    const renderer = createRenderer().use(badgePlugin);
    expect(renderer.render('::badge[success][Done]')).toContain(
      '<span class="badge badge-success">Done</span>'
    );
  });

  test('renders ::badge[error][Failed]', () => {
    const renderer = createRenderer().use(badgePlugin);
    expect(renderer.render('::badge[error][Failed]')).toContain(
      '<span class="badge badge-error">Failed</span>'
    );
  });
});

// ---------------------------------------------------------------------------
// createRenderer – unknown plugin directive
// ---------------------------------------------------------------------------

describe('createRenderer – unknown plugin directive', () => {
  test('unknown directive does not throw', () => {
    const renderer = createRenderer();
    expect(() => renderer.render('::unknown[content]')).not.toThrow();
  });
});

// ---------------------------------------------------------------------------
// createRenderer – multiple plugins
// ---------------------------------------------------------------------------

describe('createRenderer – multiple plugins', () => {
  test('multiple plugins active simultaneously', () => {
    const renderer = createRenderer()
      .use(warningPlugin)
      .use(badgePlugin);
    const input = '::warning[Attention]\n::badge[info][New]';
    const output = renderer.render(input);
    expect(output).toContain('<div class="warning">Attention</div>');
    expect(output).toContain('<span class="badge badge-info">New</span>');
  });

  test('three plugins registered in sequence and all active', () => {
    const renderer = createRenderer()
      .use(warningPlugin)
      .use(badgePlugin)
      .use(uppercasePlugin);
    const output = renderer.render('::upper[hello]');
    expect(output).toContain('<span class="upper">HELLO</span>');
  });

  test('standalone renderMarkdown does not resolve plugin directives', () => {
    const output = renderMarkdown('::warning[Test]');
    expect(output).not.toContain('<div class="warning">');
  });
});

// ---------------------------------------------------------------------------
// Integration – Markdown + Plugins
// ---------------------------------------------------------------------------

describe('Integration – Markdown + Plugins', () => {
  test('document with heading, paragraph and plugin directive', () => {
    const renderer = createRenderer().use(warningPlugin);
    const input = [
      '# Document Title',
      '',
      'Introduction with **bold** text.',
      '',
      '::warning[Please read the notes.]'
    ].join('\n');
    const output = renderer.render(input);
    expect(output).toContain('<h1>Document Title</h1>');
    expect(output).toContain('<strong>bold</strong>');
    expect(output).toContain('<div class="warning">Please read the notes.</div>');
  });

  test('plugin directive next to a list', () => {
    const renderer = createRenderer().use(badgePlugin);
    const input = '- Item A\n- Item B\n\n::badge[success][Done]';
    const output = renderer.render(input);
    expect(output).toContain('<ul>');
    expect(output).toContain('<li>Item A</li>');
    expect(output).toContain('<span class="badge badge-success">Done</span>');
  });

  test('code block protects plugin syntax from resolution', () => {
    const renderer = createRenderer().use(warningPlugin);
    const input = '```\n::warning[do not resolve]\n```';
    const output = renderer.render(input);
    expect(output).toContain('::warning[do not resolve]');
    expect(output).not.toContain('<div class="warning">');
  });

  test('complete document with all block types and two plugins', () => {
    const renderer = createRenderer().use(warningPlugin).use(badgePlugin);
    const input = [
      '# Title',
      '## Subtitle',
      '',
      'Normal paragraph with **bold** and *italic*.',
      '',
      '- List item 1',
      '- List item 2',
      '',
      '1. First',
      '2. Second',
      '',
      '> Quote',
      '',
      '---',
      '',
      '```javascript',
      'const x = 42;',
      '```',
      '',
      '::warning[Attention!]',
      '::badge[info][Status]'
    ].join('\n');
    const output = renderer.render(input);
    expect(output).toContain('<h1>Title</h1>');
    expect(output).toContain('<h2>Subtitle</h2>');
    expect(output).toContain('<strong>bold</strong>');
    expect(output).toContain('<em>italic</em>');
    expect(output).toContain('<ul>');
    expect(output).toContain('<ol>');
    expect(output).toContain('<blockquote>Quote</blockquote>');
    expect(output).toContain('<hr>');
    expect(output).toContain('class="language-javascript"');
    expect(output).toContain('<div class="warning">Attention!</div>');
    expect(output).toContain('<span class="badge badge-info">Status</span>');
  });
});
