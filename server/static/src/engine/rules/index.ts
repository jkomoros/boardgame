// The rules layer's public surface. See ../README.md.

export * from './types.ts';
export { compileGame, emptyStateJSON, gameDisplayName, newVariant, variantConfig, plainGameStackName, gameSpecError, isGameSpecError } from './compile.ts';
export type { GameSpecError } from './compile.ts';
export { newRulesState, rulesOf } from './state.ts';
export type { RulesState, TimerHandle } from './state.ts';
export type { CompiledGame, CompiledMove, ChoiceSchemaItem, Environment, InputField } from './state.ts';
export {
	createGame, propose, loadState, gameStatus, checkGameFinished, moveWithInput, moveFromRecord,
	MAX_FIX_UP_RECURSION, ERR_TOO_MANY_FIX_UPS,
} from './engine.ts';
export type { CreateOptions, CreateResult, GameSnapshot, Input, StepOptions, StepResult, VersionOutput } from './engine.ts';
export {
	viewerOutput, versionStatus, computedProperties, sanitizedMove, moveLegality, projectedMoveChoices, moveNamePublic, animationKeys,
} from './view.ts';
export type { MoveLegality, ProjectedMoveChoices, VersionSnapshot, ViewerGame, ViewerMove, ViewerOutput } from './view.ts';
export { activeTimers, activeTimersView, timerGuardError } from './timers.ts';
export type { ActiveTimer, TimerGuard, TimerMove, TimerRecord, TimerRef, TimerStatus } from './timers.ts';
export { formatInstant, parseInstant, ZERO_TIME } from './time.ts';
export * from './progression.ts';
export * from './stock.ts';
export { bindLegalState, bindLegalMove, bindLegalChest, subStateReader, valueReader, newBoundStack, PROPERTY_TYPE_NAMES } from './binding.ts';
export type { BoundStack } from './binding.ts';
export type { BindingContext, ComponentSchema, DelegateAnswers, ValuePropDef } from './binding.ts';
export { newEnumRegistry, bindEnum } from './enums.ts';
export type { EnumRegistry } from './enums.ts';
export type { BoundEnum } from './enums.ts';
