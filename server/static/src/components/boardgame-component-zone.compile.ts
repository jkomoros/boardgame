import './boardgame-component-zone.js';
import type { SelectionDraftSelectionBinding, TargetAction } from '../client.js';

const zone = document.createElement('boardgame-component-zone');
zone.label = 'Draw pile';
zone.layout = 'pile';
zone.headingLevel = 3;
zone.stack = undefined;
zone.hideCount = true;
declare const action: TargetAction<number>;
declare const numericSelection: SelectionDraftSelectionBinding<number>;
declare const stableSelection: SelectionDraftSelectionBinding<string>;
zone.action = action;
zone.action = null;
zone.selection = numericSelection;
zone.selection = stableSelection;
zone.selection = null;

// @ts-expect-error component zones deliberately exclude board geometry layouts
zone.layout = 'board';
// @ts-expect-error component-zone layouts are a closed implemented set
zone.layout = 'carousel';
// @ts-expect-error heading levels are numeric
zone.headingLevel = '2';
// @ts-expect-error labels are strings
zone.label = 12;
