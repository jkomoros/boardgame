import { test, expect } from '@playwright/test';

/**
 * The wire contract for a die's roll counter.
 *
 * `<boardgame-die>` tumbles when `item.DynamicValues.RollCount` changes, and
 * nothing else it can see says a die was thrown: a throw landing on the face
 * already showing leaves `SelectedFace` and `Value` untouched, which is about
 * one throw in six for a d6 and was one roll in six that visibly did nothing.
 *
 * The counter therefore crosses a process boundary -- `components/dice`'s Go
 * struct, the state serializer, the API, and `selectors.ts`'s component
 * expansion -- with a unit test on each side and, without this, nothing in the
 * middle. This suite is the only one that runs the API FROM SOURCE, which is
 * what makes it the right home: the parity suite talks to whatever server
 * happens to be up, so it cannot tell a missing field from a stale binary.
 */
test('the server sends a die a roll count, and increments it on every throw', async ({ page }) => {
  test.setTimeout(120_000);
  const email = 'die-roll-count@example.com';
  expect((await page.request.post('/api/auth', {
    form: { uid: email, token: 'offline-test-token', email, displayname: 'Die Roll Count' },
  })).ok()).toBe(true);

  const created = await page.request.post('/api/new/game', {
    form: { manager: 'pig', numplayers: '2' },
  });
  expect(created.ok()).toBe(true);
  const initial = await created.json() as { GameID: string; Status: string };
  expect(initial.Status).toBe('Success');

  const infoURL = `/api/game/pig/${encodeURIComponent(initial.GameID)}/info?admin=1&player=-2`;
  interface DieState {
    rollCount: unknown;
    selectedFace: unknown;
    value: unknown;
    currentPlayer: number;
  }
  const readDie = async (): Promise<DieState> => {
    const body = await (await page.request.get(infoURL)).json() as {
      Game?: {
        CurrentPlayerIndex?: number;
        CurrentState?: { Components?: Record<string, Array<Record<string, unknown>>> };
      };
    };
    const dice = body.Game?.CurrentState?.Components?.dice ?? [];
    return {
      rollCount: dice[0]?.RollCount,
      selectedFace: dice[0]?.SelectedFace,
      value: dice[0]?.Value,
      currentPlayer: body.Game?.CurrentPlayerIndex ?? 0,
    };
  };

  await expect.poll(async () => (await readDie()).selectedFace, { timeout: 30_000 })
    .toEqual(expect.any(Number));

  const first = await readDie();
  // Present at all, and a number. An API built before the counter existed sends
  // undefined here, and the die falls back to the face change -- which is the
  // behaviour this whole change exists to replace.
  expect(typeof first.rollCount,
    `the die's DynamicValues carry no RollCount: ${JSON.stringify(first)}`).toBe('number');

  const seen: number[] = [first.rollCount as number];
  let sameFaceThrows = 0;
  for (let attempt = 0; attempt < 12; attempt++) {
    const before = await readDie();
    const rolled = await page.request.post(
      `/api/game/pig/${encodeURIComponent(initial.GameID)}/move`,
      { form: { MoveType: 'Roll Dice', admin: '1', player: String(before.currentPlayer) } });
    const body = await rolled.json() as { Status?: string; Error?: string };
    expect(body.Status, body.Error).toBe('Success');

    await expect.poll(async () => (await readDie()).rollCount, { timeout: 30_000 })
      .toBe((before.rollCount as number) + 1);
    const after = await readDie();
    seen.push(after.rollCount as number);
    if (after.selectedFace === before.selectedFace) {
      sameFaceThrows++;
      // The case the counter exists for, observed over the wire: the roll
      // counter moved and NOTHING else about the die did.
      expect(after.value).toBe(before.value);
    }
    // pig's die has to be counted before it can be rolled again.
    const counted = await page.request.post(
      `/api/game/pig/${encodeURIComponent(initial.GameID)}/move`,
      { form: { MoveType: 'Count Die', admin: '1', player: String(before.currentPlayer) } });
    await counted.json();
  }

  // Strictly one per throw, never skipped and never doubled.
  expect(seen).toEqual(seen.map((_, index) => seen[0] + index));
  // Twelve throws of a d6 miss the same-face case about one run in nine, so it
  // is reported rather than asserted; a run that never saw one still proves the
  // increment.
  if (sameFaceThrows === 0) {
    console.log('note: no throw in this run landed on the face already showing');
  }
});
