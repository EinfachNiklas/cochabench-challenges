'use strict';

/**
 * Parst einen Markdown-String in ein Array von Block-Objekten.
 *
 * Unterstützte Block-Typen:
 *   - { type: 'heading1', content: string }
 *   - { type: 'heading2', content: string }
 *   - { type: 'heading3', content: string }
 *   - { type: 'code', content: string, language: string }
 *   - { type: 'ul', items: string[] }
 *   - { type: 'ol', items: string[] }
 *   - { type: 'blockquote', content: string }
 *   - { type: 'hr' }
 *   - { type: 'paragraph', content: string }
 *
 * Aufeinanderfolgende Listenzeilen desselben Typs werden zu einem Block zusammengefasst.
 *
 * @param {string} markdown
 * @returns {Array<Object>}
 */
function parseBlocks(markdown) {
  if (!markdown || !markdown.trim()) {
    return [];
  }

  const lines = markdown.split('\n');
  const blocks = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Code fence
    if (line.startsWith('```')) {
      const language = line.slice(3).trim();
      const contentLines = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) {
        contentLines.push(lines[i]);
        i++;
      }
      i++; // skip closing ```
      blocks.push({ type: 'code', content: contentLines.join('\n'), language });
      continue;
    }

    // Headings
    if (line.startsWith('### ')) {
      blocks.push({ type: 'heading3', content: line.slice(4) });
      i++;
      continue;
    }
    if (line.startsWith('## ')) {
      blocks.push({ type: 'heading2', content: line.slice(3) });
      i++;
      continue;
    }
    if (line.startsWith('# ')) {
      blocks.push({ type: 'heading1', content: line.slice(2) });
      i++;
      continue;
    }

    // Horizontal rule
    if (line.trim() === '---') {
      blocks.push({ type: 'hr' });
      i++;
      continue;
    }

    // Blockquote
    if (line.startsWith('> ')) {
      blocks.push({ type: 'blockquote', content: line.slice(2) });
      i++;
      continue;
    }

    // Unordered list — collect consecutive lines
    if (line.startsWith('- ')) {
      const items = [];
      while (i < lines.length && lines[i].startsWith('- ')) {
        items.push(lines[i].slice(2));
        i++;
      }
      blocks.push({ type: 'ul', items });
      continue;
    }

    // Ordered list — collect consecutive lines
    if (/^\d+\. /.test(line)) {
      const items = [];
      while (i < lines.length && /^\d+\. /.test(lines[i])) {
        items.push(lines[i].replace(/^\d+\. /, ''));
        i++;
      }
      blocks.push({ type: 'ol', items });
      continue;
    }

    // Empty lines (paragraph separators) — skip
    if (line.trim() === '') {
      i++;
      continue;
    }

    // Paragraph
    blocks.push({ type: 'paragraph', content: line });
    i++;
  }

  return blocks;
}

/**
 * Wendet Inline-Transformationen auf einen String an.
 *
 * Transformationen (in dieser Reihenfolge):
 *   - Inline-Code: `code` → <code>code</code>  (unterdrückt weitere Transformationen innen)
 *   - Links: [text](url) → <a href="url">text</a>
 *   - Fett: **text** → <strong>text</strong>
 *   - Kursiv: *text* → <em>text</em>
 *
 * @param {string} text
 * @returns {string}
 */
function applyInline(text) {
  if (!text) return text;

  // Inline-Code first (protects content from further transforms)
  // Use a placeholder approach to avoid transforming content inside backticks
  const codePlaceholders = [];
  let result = text.replace(/`([^`]+)`/g, (_, code) => {
    const idx = codePlaceholders.length;
    codePlaceholders.push(`<code>${code}</code>`);
    return `\x00CODE${idx}\x00`;
  });

  // Links
  result = result.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Bold (must come before italic to handle ** correctly)
  result = result.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

  // Italic
  result = result.replace(/\*([^*]+)\*/g, '<em>$1</em>');

  // Restore code placeholders
  result = result.replace(/\x00CODE(\d+)\x00/g, (_, idx) => codePlaceholders[parseInt(idx, 10)]);

  return result;
}

/**
 * Serialisiert ein Block-Array zu einem HTML-String.
 * Wendet KEINE Inline-Transformationen an — das ist Aufgabe des Integrators.
 *
 * @param {Array<Object>} blocks
 * @returns {string}
 */
function serializeBlocks(blocks) {
  return blocks.map(block => {
    switch (block.type) {
      case 'heading1': return `<h1>${block.content}</h1>`;
      case 'heading2': return `<h2>${block.content}</h2>`;
      case 'heading3': return `<h3>${block.content}</h3>`;
      case 'paragraph': return `<p>${block.content}</p>`;
      case 'blockquote': return `<blockquote>${block.content}</blockquote>`;
      case 'hr': return '<hr>';
      case 'code': {
        const cls = block.language ? ` class="language-${block.language}"` : '';
        return `<pre><code${cls}>${block.content}</code></pre>`;
      }
      case 'ul':
        return `<ul>${block.items.map(item => `<li>${item}</li>`).join('')}</ul>`;
      case 'ol':
        return `<ol>${block.items.map(item => `<li>${item}</li>`).join('')}</ol>`;
      default:
        return '';
    }
  }).join('');
}

module.exports = { parseBlocks, applyInline, serializeBlocks };
