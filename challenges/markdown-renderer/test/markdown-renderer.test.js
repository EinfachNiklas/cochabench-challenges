'use strict';

const {
  renderMarkdown,
  createRenderer,
  createPlugin
} = require('../markdown-renderer');

// ---------------------------------------------------------------------------
// Hilfs-Plugins für Tests
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
// renderMarkdown – Überschriften
// ---------------------------------------------------------------------------

describe('renderMarkdown – Überschriften', () => {
  test('rendert H1', () => {
    expect(renderMarkdown('# Hallo')).toBe('<h1>Hallo</h1>');
  });

  test('rendert H2', () => {
    expect(renderMarkdown('## Hallo')).toBe('<h2>Hallo</h2>');
  });

  test('rendert H3', () => {
    expect(renderMarkdown('### Hallo')).toBe('<h3>Hallo</h3>');
  });

  test('rendert H1 mit Inline-Fett', () => {
    expect(renderMarkdown('# Hallo **Welt**')).toBe('<h1>Hallo <strong>Welt</strong></h1>');
  });

  test('rendert H2 mit Inline-Kursiv', () => {
    expect(renderMarkdown('## Hallo *Welt*')).toBe('<h2>Hallo <em>Welt</em></h2>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Absätze
// ---------------------------------------------------------------------------

describe('renderMarkdown – Absätze', () => {
  test('rendert einfachen Absatz', () => {
    expect(renderMarkdown('Hallo Welt')).toBe('<p>Hallo Welt</p>');
  });

  test('leerer String ergibt leeren String', () => {
    expect(renderMarkdown('')).toBe('');
  });

  test('nur Whitespace ergibt leeren String', () => {
    expect(renderMarkdown('   ')).toBe('');
  });

  test('rendert mehrere Absätze', () => {
    const input = 'Erster Absatz\n\nZweiter Absatz';
    const output = renderMarkdown(input);
    expect(output).toContain('<p>Erster Absatz</p>');
    expect(output).toContain('<p>Zweiter Absatz</p>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Inline-Elemente
// ---------------------------------------------------------------------------

describe('renderMarkdown – Inline-Elemente', () => {
  test('rendert Fettschrift', () => {
    expect(renderMarkdown('**fett**')).toContain('<strong>fett</strong>');
  });

  test('rendert Kursivschrift', () => {
    expect(renderMarkdown('*kursiv*')).toContain('<em>kursiv</em>');
  });

  test('rendert Inline-Code', () => {
    expect(renderMarkdown('`code`')).toContain('<code>code</code>');
  });

  test('rendert Link', () => {
    expect(renderMarkdown('[Klick](https://example.com)')).toContain(
      '<a href="https://example.com">Klick</a>'
    );
  });

  test('kombinierte Inline-Elemente im Absatz', () => {
    const output = renderMarkdown('Text mit **fett** und *kursiv* und `code`');
    expect(output).toContain('<strong>fett</strong>');
    expect(output).toContain('<em>kursiv</em>');
    expect(output).toContain('<code>code</code>');
  });

  test('Inline-Code unterdrückt Bold-Parsing', () => {
    const output = renderMarkdown('`**kein bold**`');
    expect(output).toContain('<code>**kein bold**</code>');
    expect(output).not.toContain('<strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Code-Blöcke
// ---------------------------------------------------------------------------

describe('renderMarkdown – Code-Blöcke', () => {
  test('rendert Code-Block ohne Sprache', () => {
    const input = '```\nconsole.log("hi");\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('<pre>');
    expect(output).toContain('<code>');
    expect(output).toContain('console.log("hi");');
  });

  test('rendert Code-Block mit Sprachangabe', () => {
    const input = '```javascript\nconst x = 1;\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('class="language-javascript"');
    expect(output).toContain('const x = 1;');
  });

  test('kein Inline-Rendering innerhalb von Code-Blöcken', () => {
    const input = '```\n**nicht fett**\n```';
    const output = renderMarkdown(input);
    expect(output).toContain('**nicht fett**');
    expect(output).not.toContain('<strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Listen
// ---------------------------------------------------------------------------

describe('renderMarkdown – Listen', () => {
  test('rendert ungeordnete Liste', () => {
    const input = '- Eins\n- Zwei\n- Drei';
    const output = renderMarkdown(input);
    expect(output).toContain('<ul>');
    expect(output).toContain('<li>Eins</li>');
    expect(output).toContain('<li>Zwei</li>');
    expect(output).toContain('<li>Drei</li>');
    expect(output).toContain('</ul>');
  });

  test('rendert geordnete Liste', () => {
    const input = '1. Alpha\n2. Beta\n3. Gamma';
    const output = renderMarkdown(input);
    expect(output).toContain('<ol>');
    expect(output).toContain('<li>Alpha</li>');
    expect(output).toContain('<li>Beta</li>');
    expect(output).toContain('<li>Gamma</li>');
    expect(output).toContain('</ol>');
  });

  test('zusammenhängende Listeneinträge ergeben eine einzige Liste', () => {
    const input = '- A\n- B';
    const output = renderMarkdown(input);
    const ulCount = (output.match(/<ul>/g) || []).length;
    expect(ulCount).toBe(1);
  });

  test('Listeneinträge unterstützen Inline-Elemente', () => {
    const input = '- **fett** Element';
    const output = renderMarkdown(input);
    expect(output).toContain('<strong>fett</strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Blockzitate
// ---------------------------------------------------------------------------

describe('renderMarkdown – Blockzitate', () => {
  test('rendert einfaches Blockzitat', () => {
    expect(renderMarkdown('> Zitat')).toContain('<blockquote>Zitat</blockquote>');
  });

  test('rendert Blockzitat mit Inline-Formatierung', () => {
    const output = renderMarkdown('> **wichtig**');
    expect(output).toContain('<blockquote>');
    expect(output).toContain('<strong>wichtig</strong>');
  });
});

// ---------------------------------------------------------------------------
// renderMarkdown – Horizontale Linie
// ---------------------------------------------------------------------------

describe('renderMarkdown – Horizontale Linie', () => {
  test('rendert horizontale Linie', () => {
    expect(renderMarkdown('---')).toContain('<hr>');
  });

  test('--- im Absatz-Kontext erzeugt keine hr', () => {
    const output = renderMarkdown('Text --- Text');
    expect(output).not.toContain('<hr>');
  });
});

// ---------------------------------------------------------------------------
// createPlugin
// ---------------------------------------------------------------------------

describe('createPlugin', () => {
  test('erstellt gültiges Plugin-Objekt', () => {
    const plugin = createPlugin('test', (content) => content);
    expect(plugin).toHaveProperty('name', 'test');
    expect(plugin).toHaveProperty('handler');
    expect(typeof plugin.handler).toBe('function');
  });

  test('wirft TypeError wenn handler keine Funktion ist', () => {
    expect(() => createPlugin('test', 'kein-handler')).toThrow(TypeError);
  });

  test('wirft TypeError wenn handler null ist', () => {
    expect(() => createPlugin('test', null)).toThrow(TypeError);
  });

  test('handler wird korrekt aufgerufen', () => {
    const plugin = createPlugin('greet', (content) => `Hallo ${content}`);
    expect(plugin.handler('Welt', [])).toBe('Hallo Welt');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – Basis-Rendering
// ---------------------------------------------------------------------------

describe('createRenderer – Basis-Rendering', () => {
  test('renderer ohne Plugins verhält sich wie renderMarkdown', () => {
    const renderer = createRenderer();
    expect(renderer.render('# Titel')).toBe(renderMarkdown('# Titel'));
  });

  test('renderer gibt leeren String für leeren Input zurück', () => {
    const renderer = createRenderer();
    expect(renderer.render('')).toBe('');
  });

  test('renderer rendert Standard-Markdown korrekt', () => {
    const renderer = createRenderer();
    const output = renderer.render('**fett** und *kursiv*');
    expect(output).toContain('<strong>fett</strong>');
    expect(output).toContain('<em>kursiv</em>');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – Plugin-Registrierung
// ---------------------------------------------------------------------------

describe('createRenderer – Plugin-Registrierung', () => {
  test('use() gibt die Instanz zurück (chainable)', () => {
    const renderer = createRenderer();
    const result = renderer.use(warningPlugin);
    expect(result).toBe(renderer);
  });

  test('use() mit null wirft TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use(null)).toThrow(TypeError);
  });

  test('use() mit leerem Objekt wirft TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({})).toThrow(TypeError);
  });

  test('use() mit fehlendem handler wirft TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({ name: 'x' })).toThrow(TypeError);
  });

  test('use() mit fehlendem name wirft TypeError', () => {
    const renderer = createRenderer();
    expect(() => renderer.use({ handler: () => {} })).toThrow(TypeError);
  });

  test('Duplicate Plugin-Name überschreibt vorherige Registrierung', () => {
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
  test('rendert ::warning[...] korrekt', () => {
    const renderer = createRenderer().use(warningPlugin);
    expect(renderer.render('::warning[Achtung!]')).toContain(
      '<div class="warning">Achtung!</div>'
    );
  });

  test('::warning mit leerem Inhalt', () => {
    const renderer = createRenderer().use(warningPlugin);
    expect(renderer.render('::warning[]')).toContain('<div class="warning"></div>');
  });
});

// ---------------------------------------------------------------------------
// createRenderer – badgePlugin
// ---------------------------------------------------------------------------

describe('createRenderer – badgePlugin', () => {
  test('rendert ::badge[success][Done]', () => {
    const renderer = createRenderer().use(badgePlugin);
    expect(renderer.render('::badge[success][Done]')).toContain(
      '<span class="badge badge-success">Done</span>'
    );
  });

  test('rendert ::badge[error][Fehler]', () => {
    const renderer = createRenderer().use(badgePlugin);
    expect(renderer.render('::badge[error][Fehler]')).toContain(
      '<span class="badge badge-error">Fehler</span>'
    );
  });
});

// ---------------------------------------------------------------------------
// createRenderer – unbekannte Plugin-Direktive
// ---------------------------------------------------------------------------

describe('createRenderer – unbekannte Plugin-Direktive', () => {
  test('unbekannte Direktive führt zu keinem Fehler', () => {
    const renderer = createRenderer();
    expect(() => renderer.render('::unbekannt[inhalt]')).not.toThrow();
  });
});

// ---------------------------------------------------------------------------
// createRenderer – mehrere Plugins
// ---------------------------------------------------------------------------

describe('createRenderer – mehrere Plugins', () => {
  test('mehrere Plugins gleichzeitig aktiv', () => {
    const renderer = createRenderer()
      .use(warningPlugin)
      .use(badgePlugin);
    const input = '::warning[Achtung]\n::badge[info][Neu]';
    const output = renderer.render(input);
    expect(output).toContain('<div class="warning">Achtung</div>');
    expect(output).toContain('<span class="badge badge-info">Neu</span>');
  });

  test('drei Plugins in Folge registriert und aktiv', () => {
    const renderer = createRenderer()
      .use(warningPlugin)
      .use(badgePlugin)
      .use(uppercasePlugin);
    const output = renderer.render('::upper[hallo]');
    expect(output).toContain('<span class="upper">HALLO</span>');
  });

  test('standalone renderMarkdown löst keine Plugin-Direktiven auf', () => {
    const output = renderMarkdown('::warning[Test]');
    expect(output).not.toContain('<div class="warning">');
  });
});

// ---------------------------------------------------------------------------
// Integration – Markdown + Plugins
// ---------------------------------------------------------------------------

describe('Integration – Markdown + Plugins', () => {
  test('Dokument mit Überschrift, Absatz und Plugin-Direktive', () => {
    const renderer = createRenderer().use(warningPlugin);
    const input = [
      '# Dokumenttitel',
      '',
      'Einleitung mit **fettem** Text.',
      '',
      '::warning[Bitte lesen Sie die Hinweise.]'
    ].join('\n');
    const output = renderer.render(input);
    expect(output).toContain('<h1>Dokumenttitel</h1>');
    expect(output).toContain('<strong>fettem</strong>');
    expect(output).toContain('<div class="warning">Bitte lesen Sie die Hinweise.</div>');
  });

  test('Plugin-Direktive neben einer Liste', () => {
    const renderer = createRenderer().use(badgePlugin);
    const input = '- Eintrag A\n- Eintrag B\n\n::badge[success][Fertig]';
    const output = renderer.render(input);
    expect(output).toContain('<ul>');
    expect(output).toContain('<li>Eintrag A</li>');
    expect(output).toContain('<span class="badge badge-success">Fertig</span>');
  });

  test('Code-Block schützt Plugin-Syntax vor Auflösung', () => {
    const renderer = createRenderer().use(warningPlugin);
    const input = '```\n::warning[nicht auflösen]\n```';
    const output = renderer.render(input);
    expect(output).toContain('::warning[nicht auflösen]');
    expect(output).not.toContain('<div class="warning">');
  });

  test('vollständiges Dokument mit allen Block-Typen und zwei Plugins', () => {
    const renderer = createRenderer().use(warningPlugin).use(badgePlugin);
    const input = [
      '# Titel',
      '## Untertitel',
      '',
      'Normaler Absatz mit **fett** und *kursiv*.',
      '',
      '- Listenpunkt 1',
      '- Listenpunkt 2',
      '',
      '1. Erster',
      '2. Zweiter',
      '',
      '> Zitat',
      '',
      '---',
      '',
      '```javascript',
      'const x = 42;',
      '```',
      '',
      '::warning[Achtung!]',
      '::badge[info][Status]'
    ].join('\n');
    const output = renderer.render(input);
    expect(output).toContain('<h1>Titel</h1>');
    expect(output).toContain('<h2>Untertitel</h2>');
    expect(output).toContain('<strong>fett</strong>');
    expect(output).toContain('<em>kursiv</em>');
    expect(output).toContain('<ul>');
    expect(output).toContain('<ol>');
    expect(output).toContain('<blockquote>Zitat</blockquote>');
    expect(output).toContain('<hr>');
    expect(output).toContain('class="language-javascript"');
    expect(output).toContain('<div class="warning">Achtung!</div>');
    expect(output).toContain('<span class="badge badge-info">Status</span>');
  });
});
