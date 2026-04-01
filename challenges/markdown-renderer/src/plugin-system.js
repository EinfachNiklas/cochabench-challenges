'use strict';

/**
 * Erstellt eine neue Plugin-Registry.
 *
 * Die Registry verwaltet benannte Plugins und kann Plugin-Direktiven in einem
 * Text auflösen.
 *
 * Plugin-Direktiven-Syntax:
 *   ::pluginname[inhalt]             → handler('inhalt', [])
 *   ::pluginname[attribut][inhalt]   → handler('inhalt', ['attribut'])
 *
 * Unbekannte Direktiven (kein Plugin registriert) bleiben unverändert.
 * Duplicate-Name-Registrierung überschreibt die vorherige.
 *
 * @returns {{ register: Function, resolve: Function, list: Function }}
 */
function createRegistry() {
  const plugins = new Map();

  return {
    /**
     * Registriert ein Plugin unter dem angegebenen Namen.
     * Ein bereits vorhandener Name wird überschrieben.
     *
     * @param {string} name
     * @param {Function} handler - function(content: string, attrs: string[]) -> string
     */
    register(name, handler) {
      plugins.set(name, handler);
    },

    /**
     * Ersetzt alle Plugin-Direktiven im Text durch den jeweiligen Plugin-Output.
     * Unbekannte Direktiven bleiben unverändert.
     *
     * @param {string} text
     * @returns {string}
     */
    resolve(text) {
      if (!text) return text;

      // Match ::name[attr][content] or ::name[content]
      // The regex captures:
      //   group 1: plugin name
      //   group 2: first bracket content
      //   group 3: second bracket content (optional)
      return text.replace(/::([a-zA-Z0-9_-]+)\[([^\]]*)\](?:\[([^\]]*)\])?/g, (match, name, first, second) => {
        const handler = plugins.get(name);
        if (!handler) return match;

        if (second !== undefined) {
          // ::name[attr][content]
          return handler(second, [first]);
        } else {
          // ::name[content]
          return handler(first, []);
        }
      });
    },

    /**
     * Gibt ein Array aller registrierten Plugin-Namen zurück.
     *
     * @returns {string[]}
     */
    list() {
      return Array.from(plugins.keys());
    }
  };
}

module.exports = { createRegistry };
