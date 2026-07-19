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
      const action = (move: 'Guess Card', input: { GuessedCard: 'Guard' | 'Princess' }) => {
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
            ],
          }],
        },
        stateVersion: 5,
        schemaFingerprint: 'choices',
        schema: [{
          moveName: 'Guess Card', fieldName: 'GuessedCard', source: 'enum-values',
          candidateValues: ['Guard', 'Princess'], disclosure: 'actor-exact',
        }],
        playerPresentations: [],
        action,
        messages: { 'Guess Card': { id: 'valentine.guess', defaultMessage: 'Guess their card' } },
      });
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
    await guard.click();
    await expect.poll(() => page.evaluate(() => (
      globalThis as unknown as { __projectedProposals: unknown[] }
    ).__projectedProposals)).toEqual([expect.objectContaining({
      name: 'Guess Card',
      snapshotVersion: 5,
      arguments: { GuessedCard: 'Guard' },
    })]);

    const bounds = await surface.evaluate(element => {
      const style = getComputedStyle(element);
      return { position: style.position, maxHeight: style.maxHeight, overflow: style.overflowY };
    });
    expect(bounds.position).toBe('sticky');
    expect(bounds.maxHeight).not.toBe('none');
    expect(bounds.overflow).toBe('auto');

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
