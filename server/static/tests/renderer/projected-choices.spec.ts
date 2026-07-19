import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('generic projected choices are localized, accessible ordinary bound actions', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const setup = await page.evaluate(async () => {
      await import('/src/components/boardgame-projected-choices.ts');
      const { createMoveAction, MoveSubmissionGate } = await import('/src/moves/action.ts');
      const { buildProjectedMoveChoices } = await import('/src/moves/projected-choices.ts');
      const proposals: unknown[] = [];
      (globalThis as unknown as { __projectedProposals: unknown[] }).__projectedProposals = proposals;
      const gate = new MoveSubmissionGate();
      const service = {
        currentClientSchemaFingerprint: () => 'input',
        currentServerSchemaFingerprint: () => 'input',
        currentTransport: () => ({ submit: async (request: unknown) => {
          proposals.push(request);
          return { kind: 'success' as const };
        } }),
        currentPreviewTransport: () => null,
        currentTargetPreviewTransport: () => null,
        currentGate: () => gate,
        nextRequestID: () => 'projected-request',
        validate: () => [],
        serialize: (_move: string, input: Readonly<Record<string, unknown>>) => ({
          GuessedCard: String(input.GuessedCard),
        }),
        actionCache: new Map(),
      };
      const snapshot = {
        snapshotKey: 'v5', currentSnapshotKey: () => 'v5',
        snapshotVersion: 5, currentSnapshotVersion: () => 5,
        viewingAsPlayer: 0, proposingAsPlayer: 0, proposingAsAdmin: false,
        currentLegality: () => ({ legalForPlayer: false, legalForAnyone: true }),
        currentAnimating: () => false, baselineLegalityApplies: true,
      };
      const extraCards = [
        'Priest', 'Baron', 'Handmaid', 'Prince', 'King', 'Countess',
        'Cardinal', 'Sycophant', 'Dowager Queen', 'Constable',
      ];
      const cardValues = ['Guard', 'Princess', ...extraCards];
      const action = (move: 'Guess Card', input: { GuessedCard: string }) => {
        const builder = createMoveAction(move, service, snapshot) as ReturnType<typeof createMoveAction> & {
          with(value: typeof input): unknown;
        };
        return builder.with(input) as never;
      };
      const choices = buildProjectedMoveChoices({
        wire: {
          StateVersion: 5,
          MoveChoiceProjectionSchemaFingerprint: 'choices',
          ProjectionSchemaVersion: 1,
          Status: 'ready',
          Sets: [{
            MoveName: 'Guess Card', FieldName: 'GuessedCard', Source: 'enum-values',
            Candidates: [
              { Value: 'Guard', Available: true },
              { Value: 'Princess', Available: false },
              ...extraCards.map(Value => ({ Value, Available: true })),
            ],
          }],
        },
        stateVersion: 5,
        schemaFingerprint: 'choices',
        schema: [{
          moveName: 'Guess Card', fieldName: 'GuessedCard', source: 'enum-values',
          candidateValues: cardValues, disclosure: 'actor-exact',
        }],
        playerPresentations: [],
        action,
        messages: { 'Guess Card': { id: 'valentine.guess', defaultMessage: 'Guess their card' } },
      });
      const tallBoard = document.createElement('div');
      tallBoard.style.height = '200vh';
      tallBoard.setAttribute('aria-hidden', 'true');
      document.body.append(tallBoard);
      const underlay = document.createElement('button');
      underlay.textContent = 'Board action outside tray';
      underlay.style.cssText = 'position:fixed;left:0;bottom:0;z-index:1';
      underlay.addEventListener('click', () => { underlay.dataset.clicked = 'true'; });
      document.body.append(underlay);
      const element = document.createElement('boardgame-projected-choices');
      element.choices = choices as never;
      element.messageResolver = message => message.id === 'valentine.guess'
        ? 'Which card do they hold?'
        : message.defaultMessage;
      document.body.append(element);
      await element.updateComplete;
      return { status: choices.status, proposals };
    });
    expect(setup.status).toBe('ready');

    const surface = page.locator('boardgame-projected-choices');
    await expect(surface.locator('legend')).toHaveText('Which card do they hold?');
    const guard = surface.getByRole('button', { name: 'Which card do they hold?: Guard' });
    const princess = surface.getByRole('button', { name: 'Which card do they hold?: Princess' });
    await expect(guard).toBeEnabled();
    await expect(princess).toBeDisabled();
    const bounds = await surface.evaluate(element => {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      return {
        position: style.position,
        maxHeight: style.maxHeight,
        overflow: style.overflowY,
        top: rect.top,
        bottom: rect.bottom,
        viewportHeight: window.innerHeight,
      };
    });
    expect(bounds.position).toBe('fixed');
    expect(bounds.maxHeight).not.toBe('none');
    expect(bounds.overflow).toBe('auto');
    expect(bounds.top).toBeGreaterThanOrEqual(0);
    expect(bounds.bottom).toBeLessThanOrEqual(bounds.viewportHeight);
    const boardAction = page.getByRole('button', { name: 'Board action outside tray' });
    await boardAction.click();
    await expect(boardAction).toHaveAttribute('data-clicked', 'true');

    await page.setViewportSize({ width: 390, height: 600 });
    await page.evaluate(() => { document.documentElement.style.fontSize = '200%'; });
    const lastChoice = surface.getByRole('button', { name: 'Which card do they hold?: Constable' });
    await lastChoice.focus();
    const mobileBounds = await surface.evaluate(element => {
      const rect = element.getBoundingClientRect();
      return {
        top: rect.top,
        bottom: rect.bottom,
        viewportHeight: window.innerHeight,
        clientHeight: element.clientHeight,
        scrollHeight: element.scrollHeight,
        scrollTop: element.scrollTop,
        paddingBottom: getComputedStyle(element).paddingBottom,
      };
    });
    expect(mobileBounds.top).toBeGreaterThanOrEqual(0);
    expect(mobileBounds.bottom).toBeLessThanOrEqual(mobileBounds.viewportHeight);
    expect(mobileBounds.scrollHeight).toBeGreaterThan(mobileBounds.clientHeight);
    expect(mobileBounds.scrollTop).toBeGreaterThan(0);
    expect(mobileBounds.paddingBottom).not.toBe('');
    await expect(lastChoice).toBeVisible();

    await guard.focus();
    await guard.press('Enter');
    await expect.poll(() => page.evaluate(() => (
      globalThis as unknown as { __projectedProposals: unknown[] }
    ).__projectedProposals)).toEqual([expect.objectContaining({
      name: 'Guess Card',
      snapshotVersion: 5,
      arguments: { GuessedCard: 'Guard' },
    })]);

    const axe = await new AxeBuilder({ page })
      .include('boardgame-projected-choices')
      .withRules(['button-name', 'aria-allowed-attr'])
      .analyze();
    expect(axe.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('renderer host reserves the measured fixed tray height', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const measured = await page.evaluate(async () => {
      await import('/src/components/boardgame-render-game.ts');
      const { ProjectedMoveChoices } = await import('/src/moves/projected-choices.ts');
      const host = document.createElement('boardgame-render-game') as HTMLElement & {
        rendererLoaded: boolean;
        gameFinished: boolean;
        renderer: unknown;
        updateComplete: Promise<unknown>;
        renderRoot: ShadowRoot;
      };
      host.gameFinished = true;
      document.body.append(host);
      await host.updateComplete;
      const fakeRenderer = document.createElement('div') as HTMLDivElement & {
        choices: unknown;
        effectTheme(): object;
      };
      fakeRenderer.choices = ProjectedMoveChoices.failed();
      fakeRenderer.effectTheme = () => ({});
      host.rendererLoaded = true;
      host.renderer = fakeRenderer;
      await host.updateComplete;
      const tray = host.renderRoot.querySelector('boardgame-projected-choices') as HTMLElement & {
        updateComplete: Promise<unknown>;
      };
      await tray.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
      await host.updateComplete;
      const container = host.renderRoot.querySelector('#container');
      if (!(container instanceof HTMLElement)) throw new Error('renderer container missing');
      return {
        trayHeight: Math.ceil(tray.getBoundingClientRect().height),
        reserved: Number.parseFloat(getComputedStyle(container).paddingBottom),
      };
    });
    expect(measured.trayHeight).toBeGreaterThan(0);
    expect(measured.reserved).toBe(measured.trayHeight);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('explicit and invalid projected-choice states remain visibly failed', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-projected-choices.ts');
      const { ProjectedMoveChoices, buildProjectedMoveChoices } = await import('/src/moves/projected-choices.ts');
      let rejected = false;
      try {
        buildProjectedMoveChoices({
          wire: {
            StateVersion: 3,
            MoveChoiceProjectionSchemaFingerprint: 'choices',
            ProjectionSchemaVersion: 1,
            Status: 'ready',
            Sets: [{
              MoveName: 'Choose', FieldName: 'Value', Source: 'enum-values',
              Candidates: [{ Value: 'A', Available: false }],
            }],
          },
          stateVersion: 3,
          schemaFingerprint: 'choices',
          schema: [{
            moveName: 'Choose', fieldName: 'Value', source: 'enum-values',
            candidateValues: ['A'], disclosure: 'actor-exact',
          }],
          playerPresentations: [],
          action: () => { throw new Error('must reject before creating an action'); },
        });
      } catch {
        rejected = true;
      }
      const element = document.createElement('boardgame-projected-choices');
      element.choices = ProjectedMoveChoices.failed();
      document.body.append(element);
      await element.updateComplete;
      return rejected;
    });
    expect(result).toBe(true);
    await expect(page.getByRole('alert')).toContainText('Choices are temporarily unavailable');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
