import { sprintf } from '../../go/str.ts';
import { failT, passVerdict, unknownVerdict } from '../../legal/types.ts';
import type { LegalVerdict } from '../../legal/types.ts';
import type { GamePredicate, PredicateContext, PredicateSpec } from '../../rules/types.ts';
import { checkers_spaceIsBlack } from './bodies.gen.ts';

export function checkers_gameDelegate_ConfigurePredicateConstructors(spec: PredicateSpec): GamePredicate | string {
	if (spec.args.length !== 1) {
		return sprintf('checkers.spaceIsBlack: requires exactly 1 arg (the space-index field), got %d', spec.args.length);
	}
	const field = spec.args[0];
	let template = spec.message;
	if (template === '') {
		template = 'checkers.black_spaces_only';
	}
	return {
		name: 'checkers.spaceIsBlack',
		args: spec.args,
		reads: [{ path: field, facet: 'values' }],
		cost: 0,
		emittedTemplates: [template],
		emittedBindings: { [template]: [] },
		evaluate: (ctx: PredicateContext): LegalVerdict => {
			const [val, propType, err] = ctx.resolvePath(field);
			if (err !== null) {
				return unknownVerdict(err);
			}
			if (propType !== 'TypeEnum') {
				return unknownVerdict('checkers.spaceIsBlack: path ' + field + ' is not an enum property');
			}
			const ev = val as { value: number } | null;
			if (ev === null || typeof ev !== 'object') {
				return unknownVerdict('checkers.spaceIsBlack: path ' + field + ' resolved to a nil or non-enum value');
			}
			if (checkers_spaceIsBlack(ctx.state, ev.value)) {
				return passVerdict();
			}
			return failT(template);
		},
	};
}
