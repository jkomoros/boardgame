import '../client.js';
import { targetList, type TargetAction } from '../client.js';

declare const targets: TargetAction<'alice' | 'bob', 'Vote', { VoteTarget: number }>;
const choices = targetList(targets, key => key === 'alice' ? 'Alice' : 'Bob');
const list = document.createElement('boardgame-target-list');
list.choices = choices;
list.label = 'Vote to eliminate';
list.layout = 'grid';
list.headingLevel = 3;

// @ts-expect-error target label callbacks receive the exact key union
targetList(targets, (key: 'carol') => key);
// @ts-expect-error target lists require a controller-produced binding
list.choices = targets;
// @ts-expect-error layout is a closed policy
list.layout = 'columns';
