/**
 * Runtime validation for move proposal arguments.
 *
 * Validates arguments against the MoveForm.Fields schema already available
 * on the client. This catches wrong field names, wrong types, and missing
 * fields at runtime — the safety net for code paths that don't use the
 * framework-internal proposal events (for example gathering pickers).
 */
import type { MoveForm, MoveFormField } from '../types/api';

export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  moveName: string;
}

/**
 * Validates move arguments against the MoveForm schema.
 * Returns validation errors if any field names don't match the schema.
 *
 * Only warns — does not prevent submission. The server will reject
 * truly invalid moves via Legal() checks.
 */
export function validateMoveArgs(
  moveName: string,
  args: Record<string, unknown>,
  moveForms: MoveForm[] | null
): ValidationResult {
  const result: ValidationResult = { valid: true, errors: [], moveName };

  if (!moveForms) return result;

  // Find the move form by name
  const moveForm = moveForms.find(f => f.Name === moveName);
  if (!moveForm) {
    result.valid = false;
    result.errors.push({
      field: '*',
      message: `Move "${moveName}" not found in available moves`
    });
    return result;
  }

  if (!moveForm.Fields || moveForm.Fields.length === 0) {
    // Move has no fields — any arguments are extra
    const extraKeys = Object.keys(args).filter(k => k !== 'MoveType' && k !== 'admin' && k !== 'player');
    if (extraKeys.length > 0) {
      for (const key of extraKeys) {
        result.errors.push({
          field: key,
          message: `Field "${key}" is not defined on move "${moveName}"`
        });
      }
      result.valid = false;
    }
    return result;
  }

  // Build a set of known field names
  const knownFields = new Set(moveForm.Fields.map((f: MoveFormField) => f.Name));

  // Check for unknown fields in arguments
  for (const key of Object.keys(args)) {
    // Skip framework-internal fields
    if (key === 'MoveType' || key === 'admin' || key === 'player') continue;

    if (!knownFields.has(key)) {
      result.valid = false;
      result.errors.push({
        field: key,
        message: `Field "${key}" is not defined on move "${moveName}". Known fields: ${Array.from(knownFields).join(', ')}`
      });
    }
  }

  return result;
}

/**
 * Logs validation warnings to the console in development mode.
 * Call this from the propose-move handler.
 */
export function warnOnInvalidMoveArgs(
  moveName: string,
  args: Record<string, unknown>,
  moveForms: MoveForm[] | null
): void {
  const result = validateMoveArgs(moveName, args, moveForms);
  if (!result.valid) {
    for (const err of result.errors) {
      console.warn(`[propose-move] ${result.moveName}: ${err.message}`);
    }
  }
}
