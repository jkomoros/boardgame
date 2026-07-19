import { expect, test } from '@playwright/test';

// These tests intentionally mutate one shared in-memory Go server. Keep this
// file serial so one game's asynchronous fix-up traffic cannot perturb another
// game's transport assertions; isolated renderer fixtures remain fully parallel.
test.describe.configure({ mode: 'serial' });

test('self-started renderer server uses offline mode and same-origin API proxy', async ({ page, request }) => {
  const configResponse = await request.get('/client_config.js');
  expect(configResponse.ok()).toBe(true);
  const config = await configResponse.text();
  expect(config).toContain('"offline_dev_mode": true');
  expect(config).toContain('"host": ""');
  expect(config).toContain('"dev_host": ""');

  const apiResponse = await request.get('/api/list/manager');
  expect(apiResponse.ok()).toBe(true);
  expect(apiResponse.headers()['content-type']).toContain('application/json');

  await page.goto('/');
  await expect(page.locator('boardgame-app')).toBeAttached();
  const gameList = page.locator('boardgame-list-games-view');
  await expect(gameList).toBeAttached();
  await expect.poll(() => gameList.evaluate(element => {
    const managers = (element as unknown as { _managers?: unknown[] })._managers;
    return Array.isArray(managers) ? managers.length : 0;
  })).toBeGreaterThan(0);
  expect(await page.evaluate(() => (window as unknown as { API_HOST: string }).API_HOST)).toBe('');
});

test('offline authentication preserves encoded identity fields through the real server', async ({ page }) => {
  await page.goto('/');
  const email = 'typed+auth&test@example.com';
  const displayName = 'Ada & Grace = Players';
  await page.evaluate(({ email, displayName }) => {
    localStorage.setItem('faux-firebase-email', email);
    localStorage.setItem('faux-firebase-display-name', displayName);
  }, { email, displayName });

  const authRequestPromise = page.waitForRequest(request => (
    request.method() === 'POST' && request.url().endsWith('/api/auth')
  ));
  await page.evaluate(async () => {
    const [{ store }, { firebaseSignIn }] = await Promise.all([
      import('/src/store.ts'),
      import('/src/actions/user.ts'),
    ]);
    store.dispatch(firebaseSignIn());
  });
  const authRequest = await authRequestPromise;
  const body = new URLSearchParams(authRequest.postData() ?? '');
  expect(body.get('uid')).toBe(email);
  expect(body.get('email')).toBe(email);
  expect(body.get('displayname')).toBe(displayName);
  await expect(page.locator('boardgame-user')).toContainText(displayName);
});

test('Werewolf projected vote crosses real info, generated renderer, and move submission boundaries', async ({ page }) => {
  test.setTimeout(120_000);
  const email = 'projected-werewolf@example.com';
  const auth = await page.request.post('/api/auth', {
    form: { uid: email, token: 'offline-test-token', email, displayname: 'Voting Wolf' },
  });
  expect(auth.ok()).toBe(true);
  const created = await page.request.post('/api/new/game', {
    form: { manager: 'werewolf', numplayers: '4' },
  });
  expect(created.ok()).toBe(true);
  const creation = await created.json() as { GameID: string; GameName: string; Status: string };
  expect(creation).toMatchObject({ Status: 'Success', GameName: 'werewolf' });

  // The first info request seats the creator and the offline-dev players,
  // allowing Werewolf's gathering fix-ups to enter the first day vote.
  const infoResponse = await page.request.get(
    `/api/game/werewolf/${encodeURIComponent(creation.GameID)}/info`,
  );
  expect(infoResponse.ok()).toBe(true);
  const info = await infoResponse.json() as {
    Status: string;
    ProjectedMoveChoices?: {
      Status: string;
      Sets?: Array<{ MoveName: string; Candidates: Array<{ Value: number; Available: boolean }> }>;
    };
  };
  expect(info.Status).toBe('Success');
  const vote = info.ProjectedMoveChoices?.Sets?.find(set => set.MoveName === 'Cast Vote');
  expect(info.ProjectedMoveChoices?.Status).toBe('ready');
  expect(vote?.Candidates).toHaveLength(4);
  expect(vote?.Candidates.filter(candidate => candidate.Available)).toHaveLength(3);

  await page.goto(`/game/werewolf/${encodeURIComponent(creation.GameID)}`);
  const tray = page.locator('boardgame-projected-choices');
  await expect(tray).toBeVisible({ timeout: 20_000 });
  const enabled = tray.locator('button:not([disabled])');
  await expect(enabled).toHaveCount(3);

  const submitted = page.waitForRequest(request => (
    request.method() === 'POST'
      && request.url().includes(`/api/game/werewolf/${creation.GameID}/move`)
  ));
  await enabled.first().click();
  const request = await submitted;
  const form = new URLSearchParams(request.postData() ?? '');
  expect(form.get('MoveType')).toBe('Cast Vote');
  expect(Number(form.get('VoteTarget'))).toBeGreaterThanOrEqual(1);

  await expect.poll(async () => {
    const response = await page.request.get(
      `/api/game/werewolf/${encodeURIComponent(creation.GameID)}/info`,
    );
    const refreshed = await response.json() as {
      ProjectedMoveChoices?: { Sets?: Array<{ MoveName: string }> };
    };
    return refreshed.ProjectedMoveChoices?.Sets?.some(set => set.MoveName === 'Cast Vote') ?? false;
  }, { timeout: 20_000 }).toBe(false);
});

test('companion guest join validates room, seat options, and seat result through the real server', async ({ page, context }) => {
  test.setTimeout(120_000);
  const email = 'typed-join-host@example.com';
  await page.goto('/');
  await page.evaluate(({ email }) => {
    localStorage.setItem('faux-firebase-email', email);
    localStorage.setItem('faux-firebase-display-name', 'Join Host');
  }, { email });
  await page.evaluate(async () => {
    const [{ store }, { firebaseSignIn }] = await Promise.all([
      import('/src/store.ts'),
      import('/src/actions/user.ts'),
    ]);
    store.dispatch(firebaseSignIn());
  });
  await expect(page.locator('boardgame-user')).toContainText('Join Host');
  const createdResponse = await page.request.post('/api/new/game', {
    form: { manager: 'blackjack', numplayers: '2', companionMode: '1' },
  });
  expect(createdResponse.ok()).toBe(true);
  const created: unknown = await createdResponse.json();
  if (created === null || typeof created !== 'object' || Array.isArray(created)) {
    throw new Error('Create companion game response was not an object');
  }
  const fields = created as Readonly<Record<string, unknown>>;
  expect(fields['Status']).toBe('Success');
  expect(fields['GameName']).toBe('blackjack');
  expect(fields['GameID']).toEqual(expect.any(String));
  expect(fields['CompanionRoomCode']).toMatch(/^[A-Z]{4,5}$/);
  const roomCode = String(fields['CompanionRoomCode']);
  const gameID = String(fields['GameID']);
  // Production creation goes through actions/list.ts, which records Table
  // intent in this tab before navigation. This test creates through the raw
  // endpoint, so mirror that client transition explicitly; origin-wide cookies
  // are deliberately not enough to turn an unrelated tab into the Table.
  await page.evaluate(({ gameID }) => {
    sessionStorage.setItem(`boardgame-surface:${gameID}`, 'table');
  }, { gameID });

  const lookupResponse = await page.request.post('/api/join', { data: { code: roomCode } });
  expect(lookupResponse.ok()).toBe(true);
  const lookup = await lookupResponse.json() as { joinTicket: string };
  expect(lookup.joinTicket).toEqual(expect.any(String));
  const ticketlessOptions = await page.request.get(`/api/join/seat-options?gameID=${encodeURIComponent(gameID)}`);
  expect(ticketlessOptions.status()).toBe(401);
  expect(await ticketlessOptions.json()).toMatchObject({ code: 'JOIN_TICKET_REQUIRED' });

  const infoContract = await page.evaluate(async ({ gameID }) => {
    const response = await fetch(`/api/game/blackjack/${encodeURIComponent(gameID)}/info`);
    const payload: unknown = await response.json();
    const { decodeGameInfoResponse } = await import('/src/types/server-response.ts');
    try {
      decodeGameInfoResponse(payload);
      return { ok: true, error: '' };
    } catch (error) {
      return { ok: false, error: error instanceof Error ? error.message : String(error) };
    }
  }, { gameID: String(fields['GameID']) });
  expect(infoContract).toEqual({ ok: true, error: '' });

  await page.goto(`/game/blackjack/${encodeURIComponent(String(fields['GameID']))}`);
  const table = page.locator('boardgame-render-game-blackjack-table');
  await expect(table).toBeAttached({ timeout: 15_000 });
  const roomLock = table.getByRole('checkbox', { name: 'Lock room (no new joins)' });
  await expect(roomLock).toBeVisible();
  const lockResponse = page.waitForResponse(response => (
    response.request().method() === 'POST' && response.url().endsWith('/setRoomLock')
  ));
  await roomLock.check();
  expect((await lockResponse).ok()).toBe(true);
  await expect(roomLock).toBeChecked();
  await expect(table.locator('.host-feedback')).toContainText('Room locked');
  const lockedOptions = await page.request.get(`/api/join/seat-options?gameID=${encodeURIComponent(gameID)}`, {
    headers: { 'X-Boardgame-Join-Ticket': lookup.joinTicket },
  });
  expect(lockedOptions.status()).toBe(409);
  expect(await lockedOptions.json()).toMatchObject({ code: 'ROOM_LOCKED' });

  // Host actions are intentionally limited to one mutation per second.
  await page.waitForTimeout(1_050);
  const unlockResponse = page.waitForResponse(response => (
    response.request().method() === 'POST' && response.url().endsWith('/setRoomLock')
  ));
  await roomLock.uncheck();
  expect((await unlockResponse).ok()).toBe(true);
  await expect(roomLock).not.toBeChecked();
  await expect(table.locator('.host-feedback')).toContainText('Room unlocked');

  // Room lock is non-version metadata. Change it while this page is offline
  // and prove socket reconnect performs an authoritative info refresh rather
  // than waiting forever for a version notification that will never exist.
  await page.waitForTimeout(1_050);
  const cdp = await context.newCDPSession(page);
  await cdp.send('Network.enable');
  await cdp.send('Network.emulateNetworkConditions', {
    offline: true,
    latency: 0,
    downloadThroughput: 0,
    uploadThroughput: 0,
  });
  await expect(page.getByText('You are offline. The game will reconnect automatically when the network returns.')).toBeVisible();
  const backgroundLock = await page.request.post(
    `/api/game/blackjack/${encodeURIComponent(gameID)}/setRoomLock`,
    { data: { locked: true } },
  );
  expect(backgroundLock.ok()).toBe(true);
  await cdp.send('Network.emulateNetworkConditions', {
    offline: false,
    latency: 0,
    downloadThroughput: -1,
    uploadThroughput: -1,
  });
  await expect.poll(async () => page.evaluate(async ({ gameID }) => {
    const response = await fetch(`/api/game/blackjack/${encodeURIComponent(gameID)}/info`);
    const payload = await response.json() as { CompanionInfo?: { RoomLocked?: boolean } };
    return payload.CompanionInfo?.RoomLocked;
  }, { gameID })).toBe(true);
  await expect(roomLock).toBeChecked({ timeout: 15_000 });
  await page.waitForTimeout(1_050);
  await roomLock.uncheck();
  await expect(roomLock).not.toBeChecked();

  await page.goto(`/join?code=${encodeURIComponent(String(fields['CompanionRoomCode']))}`);
  await page.getByRole('button', { name: 'Use a new guest identity' }).click();
  await page.getByRole('button', { name: 'Looks good — join!' }).click();
  await page.waitForURL(/\/game\/blackjack\//, { timeout: 20_000 });
  await expect(page.locator('boardgame-render-game-blackjack-hand')).toBeAttached({ timeout: 15_000 });

  // A response-loss/reload retry for the same authenticated identity is
  // ordinary success and must never allocate another seat.
  const returningUID = await page.evaluate(() => localStorage.getItem('faux-firebase-email'));
  expect(returningUID).toEqual(expect.any(String));
  const retryLookupResponse = await page.request.post('/api/join', { data: { code: roomCode } });
  expect(retryLookupResponse.ok()).toBe(true);
  const retryLookup = await retryLookupResponse.json() as { joinTicket: string };
  const retryClaim = await page.request.post('/api/join/seat', {
    headers: { 'X-Boardgame-Join-Ticket': retryLookup.joinTicket },
    data: {
      gameID,
      uid: returningUID,
      displayName: 'ReturningFox',
      avatarSlug: '🦊',
      seatPick: -1,
      attemptID: 'lost-response-retry',
    },
  });
  expect(retryClaim.ok()).toBe(true);
  expect(await retryClaim.json()).toMatchObject({ gameID, resumed: true });

	// The original Table socket disappeared when this browser entered the
	// join flow. Even though its HttpOnly lease cookie still exists, the Hand
	// socket must not renew it. After the grace period, framework-owned recovery
	// appears and promotes exactly this device back to the Table.
	const takeover = page.getByRole('button', { name: 'Take over shared Table' });
	await expect(takeover).toBeVisible({ timeout: 50_000 });
	await takeover.click();
	await expect(page.locator('boardgame-render-game-blackjack-table')).toBeAttached({ timeout: 15_000 });
	await expect(page.getByRole('checkbox', { name: 'Lock room (no new joins)' })).toBeVisible();
});

test('an active shared Table transfers atomically to a fresh accountless screen', async ({ page, browser }) => {
  test.setTimeout(120_000);
  const email = 'table-transfer-host@example.com';
  const auth = await page.request.post('/api/auth', {
    form: { uid: email, token: 'offline-test-token', email, displayname: 'Transfer Host' },
  });
  expect(auth.ok()).toBe(true);
  const created = await page.request.post('/api/new/game', {
    form: { manager: 'blackjack', numplayers: '2', companionMode: '1' },
  });
  expect(created.ok()).toBe(true);
  const result = await created.json() as { GameID: string; GameName: string; Status: string; CompanionRoomCode: string };
  expect(result.Status).toBe('Success');

  // Seat the same identity from an isolated phone context. The source browser
  // remains a Table but now also carries a seated user's auth identity, making
  // the post-fence privacy assertion below meaningful.
  const seatContext = await browser.newContext();
  const seatPage = await seatContext.newPage();
  try {
    expect((await seatPage.request.post('/api/auth', {
      form: { uid: email, token: 'offline-test-token', email, displayname: 'Transfer Host' },
    })).ok()).toBe(true);
    const lookup = await seatPage.request.post('/api/join', { data: { code: result.CompanionRoomCode } });
    expect(lookup.ok()).toBe(true);
    const { joinTicket } = await lookup.json() as { joinTicket: string };
    const claim = await seatPage.request.post('/api/join/seat', {
      headers: { 'X-Boardgame-Join-Ticket': joinTicket },
      data: { gameID: result.GameID, uid: email, displayName: 'TransferHost', avatarSlug: '🦊', seatPick: -1, attemptID: 'transfer-privacy-seat' },
    });
    expect(claim.ok()).toBe(true);
  } finally {
    await seatContext.close();
  }

  // The API call above bypasses the production create-game navigation, which
  // carries the new Table tab's explicit presentation intent in the URL.
  await page.goto(`/game/blackjack/${encodeURIComponent(result.GameID)}?display=table`);
  await expect(page.locator('boardgame-render-game-blackjack-table')).toBeAttached({ timeout: 15_000 });
  await page.getByRole('button', { name: 'Move shared Table' }).click();
  const dialog = page.getByRole('dialog', { name: 'Move the shared Table' });
  await expect(dialog).toBeVisible();
  const cancelledURL = (await dialog.locator('.table-transfer-url').textContent())?.trim();
  await dialog.getByRole('button', { name: 'Cancel transfer' }).click();
  await expect(dialog).toBeHidden();
  await page.getByRole('button', { name: 'Move shared Table' }).click();
  await expect(dialog).toBeVisible();
  const claimURL = (await dialog.locator('.table-transfer-url').textContent())?.trim();
  const manualCode = (await dialog.locator('.table-transfer-code').textContent())?.trim();
  expect(claimURL).toMatch(/^http:\/\/127\.0\.0\.1:\d+\/table#transfer=/);
  expect(claimURL).not.toBe(cancelledURL);
  expect(manualCode).toMatch(/^[0-9A-HJKMNP-TV-Z]{10}$/);

  const receiverContext = await browser.newContext();
  const receiver = await receiverContext.newPage();
  try {
    await receiver.goto(cancelledURL!);
    await expect(receiver.getByRole('alert')).toContainText('not valid');
    await receiver.goto(new URL('/table', claimURL!).toString());
    await receiver.getByLabel('Room code').fill(result.CompanionRoomCode);
    await receiver.getByLabel('Transfer code').fill(manualCode!);
    await receiver.getByRole('button', { name: 'Continue' }).click();
    await expect(receiver.getByRole('heading', { name: 'Move Blackjack here?' })).toBeVisible();
    await receiver.getByRole('button', { name: 'Use a different code' }).click();

    await receiver.goto(claimURL!);
    await expect(receiver).toHaveURL(/\/table$/);
    await expect(receiver.getByRole('heading', { name: 'Move Blackjack here?' })).toBeVisible();
    expect((await receiverContext.cookies()).some(cookie => cookie.name === 'c')).toBe(false);

    // Let the server commit the handoff, then deliberately lose the first
    // response. Reloading /table must restore the scrubbed bearer from
    // sessionStorage, reuse the same device ID, and redeem idempotently.
    let droppedCommittedResponse = false;
    await receiver.route('**/api/table-transfer/redeem', async route => {
      if (droppedCommittedResponse) {
        await route.continue();
        return;
      }
      droppedCommittedResponse = true;
      await route.fetch();
      await route.abort('failed');
    });
    await receiver.getByRole('button', { name: 'Make this the shared Table' }).click();
    await expect.poll(() => droppedCommittedResponse).toBe(true);
    await receiver.reload();
    await receiver.waitForURL(new RegExp(`/game/blackjack/${result.GameID}\\?display=table$`), { timeout: 20_000 });
    await expect(receiver.locator('boardgame-render-game-blackjack-table')).toBeAttached({ timeout: 15_000 });
    await expect(receiver.getByRole('checkbox', { name: 'Lock room (no new joins)' })).toBeVisible();
    await receiver.unroute('**/api/table-transfer/redeem');

    // The capability, rather than a login, grants framework host controls.
    const lockResponse = receiver.waitForResponse(response => (
      response.request().method() === 'POST' && response.url().endsWith('/setRoomLock')
    ));
    await receiver.getByRole('checkbox', { name: 'Lock room (no new joins)' }).check();
    expect((await lockResponse).ok()).toBe(true);

    // The old Table is fenced immediately and receives an intentional,
    // terminal handoff state rather than continuing as a second controller.
    await expect(page.getByRole('heading', { name: 'The shared Table moved successfully' })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('The game is now running on the new screen.')).toBeVisible();
    const displacedPerspective = await page.evaluate(async ({ gameID }) => {
      const response = await fetch(`/api/game/blackjack/${encodeURIComponent(gameID)}/info?admin=1&player=0&autoCurrentPlayer=1`);
      const body = await response.json() as Readonly<Record<string, unknown>>;
      return body['ViewingAsPlayer'];
    }, { gameID: result.GameID });
    expect(displacedPerspective).toBe(-1);

    const replayContext = await browser.newContext();
    const replay = await replayContext.newPage();
    try {
      await replay.goto(claimURL!);
      await expect(replay.getByRole('heading', { name: 'Reconnect Blackjack here?' })).toBeVisible();
      await replay.getByRole('button', { name: 'Reconnect this shared Table' }).click();
      await expect(replay.getByRole('alert')).toContainText('already used by another screen');
      await expect(receiver.locator('boardgame-render-game-blackjack-table')).toBeAttached();
    } finally {
      await replayContext.close();
    }
  } finally {
    await receiverContext.close();
  }
});

test('a finished companion room rematches once and carries every surface and seat forward', async ({ page, browser }) => {
  test.setTimeout(150_000);
  const ownerEmail = 'rematch-owner@example.com';
  expect((await page.request.post('/api/auth', {
    form: { uid: ownerEmail, token: 'offline-test-token', email: ownerEmail, displayname: 'Rematch Owner' },
  })).ok()).toBe(true);
  const created = await page.request.post('/api/new/game', {
    form: {
      manager: 'blackjack', numplayers: '2', companionMode: '1', variant_maxrounds: '1',
    },
  });
  expect(created.ok()).toBe(true);
  const initial = await created.json() as {
    GameID: string; GameName: string; Status: string; CompanionRoomCode: string;
  };
  expect(initial.Status).toBe('Success');

  const joinPlayer = async (email: string, displayName: string, avatarSlug: string) => {
    const context = await browser.newContext();
    const hand = await context.newPage();
    expect((await hand.request.post('/api/auth', {
      form: { uid: email, token: 'offline-test-token', email, displayname: displayName },
    })).ok()).toBe(true);
    const lookup = await hand.request.post('/api/join', { data: { code: initial.CompanionRoomCode } });
    expect(lookup.ok()).toBe(true);
    const { joinTicket } = await lookup.json() as { joinTicket: string };
    const claim = await hand.request.post('/api/join/seat', {
      headers: { 'X-Boardgame-Join-Ticket': joinTicket },
      data: {
        gameID: initial.GameID, uid: email, displayName, avatarSlug,
        seatPick: -1, attemptID: `rematch-${email}`,
      },
    });
    expect(claim.ok()).toBe(true);
    await hand.goto(`/game/blackjack/${encodeURIComponent(initial.GameID)}?display=hand`);
    await expect(hand.locator('boardgame-render-game-blackjack-hand')).toBeAttached({ timeout: 15_000 });
    return { context, hand };
  };

  const first = await joinPlayer('rematch-one@example.com', 'Bright Fox', '🦊');
  const second = await joinPlayer('rematch-two@example.com', 'Calm Owl', '🦉');
  const adminContext = await browser.newContext();
  const admin = await adminContext.newPage();
  try {
    expect((await admin.request.post('/api/auth', {
      form: { uid: ownerEmail, token: 'offline-test-token', email: ownerEmail, displayname: 'Rematch Owner' },
    })).ok()).toBe(true);
    await page.goto(`/game/blackjack/${encodeURIComponent(initial.GameID)}?display=table`);
    await expect(page.locator('boardgame-render-game-blackjack-table')).toBeAttached({ timeout: 15_000 });

    const infoURL = `/api/game/blackjack/${encodeURIComponent(initial.GameID)}/info?admin=1&player=-2`;
    await expect.poll(async () => {
      const response = await admin.request.get(infoURL);
      const body = await response.json() as { Forms?: Array<{ Name?: string; LegalForPlayer?: boolean }> };
      // Presence alone is not enough: the move is registered for the whole
      // game, but it only becomes legal for the admin once normal play begins.
      return body.Forms?.some(form => form.Name === 'Force Finish Turn' && form.LegalForPlayer === true) ?? false;
    }, { timeout: 20_000 }).toBe(true);

    for (let attempt = 0; attempt < 12; attempt++) {
      const current = await (await admin.request.get(infoURL)).json() as {
        Game?: { Finished?: boolean; CurrentPlayerIndex?: number };
      };
      if (current.Game?.Finished) break;
      const currentPlayer = current.Game?.CurrentPlayerIndex;
      expect(currentPlayer).toEqual(expect.any(Number));
      const stood = await admin.request.post(`/api/game/blackjack/${encodeURIComponent(initial.GameID)}/move`, {
        form: { MoveType: 'Current Player Stand', admin: '1', player: String(currentPlayer) },
      });
      const stoodBody = await stood.json() as { Status?: string; Error?: string };
      expect(stoodBody.Status, stoodBody.Error).toBe('Success');
      await page.waitForTimeout(250);
    }
    const finalInitial = await (await admin.request.get(infoURL)).json() as {
      Game?: { Finished?: boolean; Version?: number; CurrentState?: { Game?: Readonly<Record<string, unknown>> } };
    };
    expect(finalInitial.Game?.Finished, JSON.stringify(finalInitial.Game)).toBe(true);

    const playAgain = page.getByRole('button', { name: 'Play again with the same players' });
    await expect(playAgain).toBeVisible({ timeout: 15_000 });
    await playAgain.click();
    await page.waitForURL(url => url.pathname.includes('/game/blackjack/') && !url.pathname.includes(initial.GameID), { timeout: 25_000 });
    const rematchID = new URL(page.url()).pathname.split('/').filter(Boolean).at(-1)!;
    expect(rematchID).not.toBe(initial.GameID);
    await expect(page.locator('boardgame-render-game-blackjack-table')).toBeAttached({ timeout: 15_000 });

    await expect.poll(() => new URL(first.hand.url()).pathname, { timeout: 20_000 })
      .toContain(`/game/blackjack/${rematchID}`);
    await expect.poll(() => new URL(second.hand.url()).pathname, { timeout: 20_000 })
      .toContain(`/game/blackjack/${rematchID}`);
    await expect(first.hand.locator('boardgame-render-game-blackjack-hand')).toBeAttached();
    await expect(second.hand.locator('boardgame-render-game-blackjack-hand')).toBeAttached();

    const rematchInfo = await (await first.hand.request.get(
      `/api/game/blackjack/${encodeURIComponent(rematchID)}/info`,
    )).json() as {
      ViewingAsPlayer?: number;
      CompanionInfo?: { RoomCode?: string; SeatPresentations?: Array<{ displayName: string }> };
    };
    expect(rematchInfo.ViewingAsPlayer).toBe(0);
    expect(rematchInfo.CompanionInfo?.RoomCode).not.toBe(initial.CompanionRoomCode);
    expect(rematchInfo.CompanionInfo?.SeatPresentations?.map(seat => seat.displayName)).toEqual([
      'Bright Fox', 'Calm Owl',
    ]);

    // Lost responses/retries converge on the already-published successor.
    const retry = await page.request.post(`/api/game/blackjack/${encodeURIComponent(initial.GameID)}/rematch`, { data: {} });
    expect(retry.ok()).toBe(true);
    expect(await retry.json()).toMatchObject({ ok: true, gameID: rematchID });
  } finally {
    await adminContext.close();
    await first.context.close();
    await second.context.close();
  }
});

test('assembled Pig renderer reports a real authoritative rejection through typed transport', async ({ page }) => {
  const email = 'typed-action@example.com';
  const auth = await page.request.post('/api/auth', {
    form: {
      uid: email,
      token: 'offline-test-token',
      email,
      displayname: 'Typed Action',
    },
  });
  expect(auth.ok()).toBe(true);

  const created = await page.request.post('/api/new/game', {
    form: { manager: 'pig', numplayers: '2' },
  });
  expect(created.ok()).toBe(true);
  const result = await created.json() as { GameID: string; Status: string };
  expect(result.Status).toBe('Success');

  const failedRendererRequests: string[] = [];
  page.on('requestfailed', (failed) => {
    if (failed.url().includes('/game-src/pig/') || failed.url().endsWith('/src/client.js')) {
      failedRendererRequests.push(`${failed.url()}: ${failed.failure()?.errorText || 'unknown failure'}`);
    }
  });
  await page.goto(`/game/pig/${result.GameID}`);
  const renderer = page.locator('boardgame-render-game-pig');
  await expect(renderer).toBeAttached({ timeout: 15_000 });
  // Let Pig's initial fix-up chain settle before installing the presentation
  // perspective used by this transport-focused test.
  await expect.poll(() => renderer.evaluate(element => (
    element as unknown as { gameVersion: number }
  ).gameVersion)).toBeGreaterThanOrEqual(2);
  // This server-backed smoke game intentionally has no occupied seats. Make
  // only the presentation verdict available so the test exercises the action
  // and HTTP transport; server legality remains authoritative and may reject.
  const moveRequestPromise = page.waitForRequest(candidate =>
    candidate.method() === 'POST' && candidate.url().endsWith(`/api/game/pig/${result.GameID}/move`));
  // Install and consume the test perspective atomically so a websocket render
  // cannot replace it between setup and action creation.
  const proposalPromise = renderer.evaluate(element => {
    const typed = element as unknown as {
      currentPlayerIndex: number;
      gameVersion: number;
      viewingAsPlayer: number;
      proposingAsPlayer: number;
      proposingAsAdmin: boolean;
      animating: boolean;
      moveLegality: Record<string, { legalForPlayer: boolean; legalForAnyone: boolean }>;
      move(name: 'Roll Dice'): { propose(): Promise<unknown> };
    };
    typed.viewingAsPlayer = (typed.currentPlayerIndex + 1) % 2;
    typed.proposingAsPlayer = typed.viewingAsPlayer;
    // Admin perspective makes the requested non-current player authoritative.
    // Without it the server correctly ignores `player` and uses this user's
    // seated identity, which can make the supposedly illegal move succeed.
    typed.proposingAsAdmin = true;
    // Pig's debug player continuously advances the game and can keep the
    // renderer animation gate closed. This test targets HTTP submission, so
    // consume the current presentation atomically with that local gate open.
    typed.animating = false;
    typed.moveLegality = {
      'Roll Dice': { legalForPlayer: true, legalForAnyone: true },
      'Done Turn': { legalForPlayer: false, legalForAnyone: true },
    };
    return typed.move('Roll Dice').propose();
  });
  const moveRequest = await moveRequestPromise;
  const body = new URLSearchParams(moveRequest.postData() ?? '');
  expect(body.get('MoveType')).toBe('Roll Dice');
  expect(body.get('player')).toMatch(/^[01]$/);
  expect(body.get('admin')).toBe('1');
  expect(body.get('ExpectedVersion')).toMatch(/^\d+$/);

  expect((await moveRequest.response())?.ok()).toBe(true);
  const proposal = await proposalPromise as { kind: string; requestID: string };
  // Normally this is an illegal-proposer rejection. If the debug player moves
  // concurrently, ExpectedVersion instead produces a stale-snapshot rejection.
  // Both prove the real server remained authoritative; success never does.
  expect(['server-rejection', 'stale-snapshot']).toContain(proposal.kind);
  expect(proposal.requestID).toEqual(expect.any(String));
  expect(failedRendererRequests).toEqual([]);
});

test('legacy declarative renderer controls are inert', async ({ page }) => {
  await page.goto('/');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-base-game-renderer.ts');
    const renderer = document.createElement('boardgame-base-game-renderer') as HTMLElement & {
      updateComplete: Promise<unknown>;
    };
    let proposals = 0;
    renderer.addEventListener('propose-move', () => proposals++);
    document.body.append(renderer);
    await renderer.updateComplete;

    const button = document.createElement('button');
    button.setAttribute('propose-move', 'Choose');
    renderer.append(button);
    button.click();
    await Promise.resolve();
    const hasProposalShortcut = 'proposeMove' in renderer;
    renderer.remove();
    return { proposals, hasProposalShortcut };
  });

  expect(result).toEqual({ proposals: 0, hasProposalShortcut: false });
});
