/**
 * Converts a dash-case string to camelCase.
 * Example: "component-index" -> "componentIndex"
 */
export function dashToCamelCase(dash: string): string {
  return dash.replace(/-([a-z])/g, (m) => m[1].toUpperCase());
}
