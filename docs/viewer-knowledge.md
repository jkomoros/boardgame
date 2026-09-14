# Reading viewer knowledge

The server sanitizes each player's snapshot before sending it. For compatibility,
hidden scalar fields still contain their zero/default value. That value alone
does not establish what the viewer knows.

Use the typed facade when a property can be hidden:

```ts
import { viewPlayerProp, viewGameProp } from '../../src/client.js';

const role = viewPlayerProp(this.state, this.viewingAs, 'Role');
const label = role.known ? role.value : 'Unknown role';
const score = viewGameProp(this.state, 'SecretScore');
```

Keys and results retain the generated state's types. Known zero, false, and
default enum values are ordinary values. An unknown result has no `value` field.
Missing players, properties, or visibility metadata return unknown. Legacy
fixtures therefore need real viewer snapshots to demonstrate known values; do
not synthesize knowledge by inspecting zero values. Existing direct access to
public state is unchanged.

`Game.JSONForPlayer` includes `CurrentState.Visibility`. Its `Game` and `Players`
maps give each property the facets that survived sanitization: `values`,
`count`, `occupancy`, `order`, and `nonempty`. The server derives these from the
same truth table used for declarative legality. Clients can use
`facetAvailable(facets, 'count')` when they only need a stack's count. This
metadata is part of the viewer projection, not persistent game state. Plain
state marshaling and golden records keep their existing shape.

Game/player properties are covered in this first version. Dynamic component
properties and custom computed properties are not covered. Their disclosure
requires additional effective per-component policies or deliberate game-owned
projections; do not borrow the containing stack's visibility for those values.
Framework `RoleValue`, `TeamValue`, and `ColorValue` labels are omitted when the
source field is unavailable. Go computed presentation code can use
`boardgame.PropertyFacetAvailable(player, "Role", boardgame.LegalFacetValues)`.

Visibility decisions themselves are now observable. Custom sanitization policy
must treat the distinction between known and unknown as intended disclosure;
it must not encode an otherwise secret fact in that decision. Hidden values
remain sanitized on the server. A client-side availability check is a display
aid, never an authorization boundary.
