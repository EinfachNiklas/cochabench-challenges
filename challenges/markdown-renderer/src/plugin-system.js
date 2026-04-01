'use strict';

/**
 * Creates a new plugin registry.
 *
 * The registry manages named plugins and can resolve plugin directives in a
 * text string.
 *
 * Plugin directive syntax:
 *   ::pluginname[content]            → handler('content', [])
 *   ::pluginname[attribute][content] → handler('content', ['attribute'])
 *
 * Unknown directives (no plugin registered) are left unchanged.
 * Registering a duplicate name overwrites the previous entry.
 *
 * @returns {{ register: Function, resolve: Function, list: Function }}
 */
function createRegistry() {
  const plugins = new Map();

  return {
    /**
     * Registers a plugin under the given name.
     * An existing entry with the same name is overwritten.
     *
     * @param {string} name
     * @param {Function} handler - function(content: string, attrs: string[]) -> string
     */
    register(name, handler) {
      plugins.set(name, handler);
    },

    /**
     * Replaces all plugin directives in the text with the respective plugin output.
     * Unknown directives are left unchanged.
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
     * Returns an array of all registered plugin names.
     *
     * @returns {string[]}
     */
    list() {
      return Array.from(plugins.keys());
    }
  };
}

module.exports = { createRegistry };
