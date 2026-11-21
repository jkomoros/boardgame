// Template keys and their default English text: the TypeScript mirror of
// Go's legal/templates.go and the Template* constants spread across the
// catalog files. A failure carries a KEY plus bindings; text is looked up at
// render time in a table shaped like the chest's LegalTemplates.

import { LEGAL_ANY_FAILED_TEMPLATE } from './predicate.ts';

export const TEMPLATE_PROP_AT_LEAST = 'legal.prop_at_least';
export const TEMPLATE_PROP_COMPARE = 'legal.prop_compare';
export const TEMPLATE_PLAYER_BOOL = 'legal.player_bool';
export const TEMPLATE_PLAYER_ALREADY_SUBMITTED = 'legal.player_already_submitted';
export const TEMPLATE_PLAYER_NOT_SUBMITTED = 'legal.player_not_submitted';
export const TEMPLATE_PLAYER_INACTIVE = 'legal.player_inactive';
export const TEMPLATE_PLAYER_ACTIVE = 'legal.player_active';
export const TEMPLATE_SEAT_NOT_FILLED = 'legal.seat_not_filled';
export const TEMPLATE_SEAT_NOT_CLOSED = 'legal.seat_not_closed';
export const TEMPLATE_PLAYER_NOT_ADMIN = 'legal.player_not_admin';
export const TEMPLATE_COMPONENT_MISSING = 'legal.component_missing';
export const TEMPLATE_COMPONENT_PRESENT_UNEXPECTED = 'legal.component_present_unexpected';
export const TEMPLATE_COMPONENT_MISSING_KEY = 'legal.component_missing_key';
export const TEMPLATE_NO_COMPONENT_TO_MOVE = 'legal.no_component_to_move';
export const TEMPLATE_MAY_NOT_MOVE_TO = 'legal.may_not_move_to';
export const TEMPLATE_MAY_NOT_MOVE_ALL_TO = 'legal.may_not_move_all_to';
export const TEMPLATE_MAY_NOT_MOVE_COUNT_TO = 'legal.may_not_move_count_to';
export const TEMPLATE_MAY_NOT_SWAP_COMPONENTS = 'legal.may_not_swap_components';
export const TEMPLATE_ALL_ACTIVE_PLAYERS = 'legal.all_active_players_failed';
export const TEMPLATE_PROPOSER_TARGET_INVALID = 'legal.proposer_target_invalid';
export const TEMPLATE_PROPOSER_NOT_YOUR_TURN = 'legal.proposer_not_your_turn';
export const TEMPLATE_PROPOSER_NOT_MOVE_PLAYER = 'legal.proposer_not_move_player';
export const TEMPLATE_NO_CARD_HERE = 'legal.no_card_here';
export const TEMPLATE_ALREADY_REVEALED = 'legal.already_revealed';
export const TEMPLATE_COMPONENT_PROP_NOT_CURRENT_PLAYER = 'legal.component_prop_not_current_player';
export const TEMPLATE_IN_PHASE = 'legal.in_phase';
export const TEMPLATE_IN_PROGRESSION = 'legal.in_progression';
export const TEMPLATE_STACK_CONSTRAINTS = 'legal.stack_constraints';
export const TEMPLATE_STACK_COUNT = 'legal.stack_count';
export const TEMPLATE_STACK_EMPTY = 'legal.stack_empty';
export const TEMPLATE_STACK_NOT_EMPTY = 'legal.stack_not_empty';
export const TEMPLATE_PROP_EQUALS = 'legal.prop_equals';
export const TEMPLATE_PROP_NOT_EQUALS = 'legal.prop_not_equals';
export const TEMPLATE_COMPONENT_PROP_EQUALS = 'legal.component_prop_equals';
export const TEMPLATE_COMPONENT_PROP_NOT_EQUALS = 'legal.component_prop_not_equals';
export const TEMPLATE_COMPONENT_PROPS_EQUAL = 'legal.component_props_equal';
export const TEMPLATE_COMPONENT_PROPS_NOT_EQUAL = 'legal.component_props_not_equal';

const DEFAULT_TEMPLATES: Readonly<Record<string, string>> = {
  [TEMPLATE_PROP_AT_LEAST]: 'requires at least {min}, but the current value is {value}',
  [TEMPLATE_PROP_COMPARE]: 'requires the current value ({value}) to be {op} {n}',
  [TEMPLATE_PLAYER_BOOL]: 'requires {prop} to be {want}',
  [TEMPLATE_PLAYER_ALREADY_SUBMITTED]: 'that player has already submitted',
  [TEMPLATE_PLAYER_NOT_SUBMITTED]: 'that player has not submitted',
  [TEMPLATE_PLAYER_INACTIVE]: 'that player is inactive',
  [TEMPLATE_PLAYER_ACTIVE]: 'that player is active',
  [TEMPLATE_SEAT_NOT_FILLED]: 'that seat is not filled',
  [TEMPLATE_SEAT_NOT_CLOSED]: 'that seat is not closed',
  [TEMPLATE_PLAYER_NOT_ADMIN]: 'that player is not a game administrator',
  [TEMPLATE_COMPONENT_MISSING]: 'there is no component at index {index}',
  [TEMPLATE_COMPONENT_PRESENT_UNEXPECTED]: 'there is unexpectedly a component at index {index}',
  [TEMPLATE_COMPONENT_MISSING_KEY]: 'there is no component at key {key}',
  [TEMPLATE_NO_COMPONENT_TO_MOVE]: 'there is no component at index {index} to move',
  [TEMPLATE_MAY_NOT_MOVE_TO]: '{detail}',
  [TEMPLATE_MAY_NOT_MOVE_ALL_TO]: '{detail}',
  [TEMPLATE_MAY_NOT_MOVE_COUNT_TO]: '{detail}',
  [TEMPLATE_MAY_NOT_SWAP_COMPONENTS]: '{detail}',
  [TEMPLATE_ALL_ACTIVE_PLAYERS]: 'not every active player satisfies the required condition',
  [TEMPLATE_PROPOSER_TARGET_INVALID]: '{detail}',
  [TEMPLATE_PROPOSER_NOT_YOUR_TURN]: '{detail}',
  [TEMPLATE_PROPOSER_NOT_MOVE_PLAYER]: "the move's player does not match the proposer",
  [TEMPLATE_NO_CARD_HERE]: 'there is no card at that index',
  [TEMPLATE_ALREADY_REVEALED]: 'that card has already been revealed',
  [TEMPLATE_COMPONENT_PROP_NOT_CURRENT_PLAYER]: "the component's {prop} does not match the current player's {prop}",
  [TEMPLATE_IN_PHASE]: '{detail}',
  [TEMPLATE_IN_PROGRESSION]: '{detail}',
  [TEMPLATE_STACK_CONSTRAINTS]: '{detail}',
  [TEMPLATE_STACK_COUNT]: "requires the stack's count ({count}) to be {op} {n}",
  [TEMPLATE_STACK_EMPTY]: 'requires the stack to be empty',
  [TEMPLATE_STACK_NOT_EMPTY]: 'requires the stack to not be empty',
  [TEMPLATE_PROP_EQUALS]: 'requires the current value ({value}) to equal {want}',
  [TEMPLATE_PROP_NOT_EQUALS]: 'requires the current value ({value}) to not equal {want}',
  [TEMPLATE_COMPONENT_PROP_EQUALS]: "requires the component's {prop} ({value}) to equal {want}",
  [TEMPLATE_COMPONENT_PROP_NOT_EQUALS]: "requires the component's {prop} ({value}) to not equal {want}",
  [TEMPLATE_COMPONENT_PROPS_EQUAL]: "requires the selected components' {prop} values ({left} and {right}) to match",
  [TEMPLATE_COMPONENT_PROPS_NOT_EQUAL]: "requires the selected components' {prop} values ({left} and {right}) to differ",
  [LEGAL_ANY_FAILED_TEMPLATE]: 'none of the required conditions were satisfied',
};

/**
 * Go: legal.DefaultTemplates(). A fresh copy of the default English table.
 * A game's table is this overlaid with its own entries (mergeLegalTemplates),
 * which is what the chest ships as LegalTemplates.
 */
export function defaultLegalTemplates(): Record<string, string> {
  return { ...DEFAULT_TEMPLATES };
}

/**
 * Go: buildLegalRegistryAndTemplates' template half. The defaults overlaid
 * with a game's own additions and overrides; a later table wins per key.
 */
export function mergeLegalTemplates(...tables: readonly (Readonly<Record<string, string>> | null | undefined)[]): Record<string, string> {
  const out = defaultLegalTemplates();
  for (const table of tables) {
    if (table) {
      Object.assign(out, table);
    }
  }
  return out;
}
