import '../client.js';
import type { InspectorOpenChangedDetail } from '../client.js';

const inspector = document.createElement('boardgame-inspector');
inspector.label = 'Vision card';
inspector.description = 'A larger view of the selected vision card';
inspector.triggerLabel = 'Inspect vision card';
inspector.dismissible = false;
inspector.open = true;
inspector.show();
inspector.close('programmatic');
inspector.addEventListener('inspector-open-changed', event => {
  const detail: InspectorOpenChangedDetail = event.detail;
  detail.reason satisfies 'backdrop' | 'close-button' | 'escape' | 'programmatic' | 'trigger';
});

// @ts-expect-error open state is boolean
inspector.open = 'yes';
// @ts-expect-error callers cannot forge a browser Escape reason
inspector.close('escape');
