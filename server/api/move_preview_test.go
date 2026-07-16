package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	blackjack "github.com/jkomoros/boardgame/examples/blackjack"
	tictactoe "github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/sirupsen/logrus"
)

func TestMoveFormBindingPreservesContextDefaultsButRejectsMissingRequiredInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{}

	blackjackManager, err := boardgame.NewGameManager(blackjack.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("building blackjack manager: %v", err)
	}
	blackjackGame, err := blackjackManager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building blackjack game: %v", err)
	}
	request := func(moveType string) *gin.Context {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		values := url.Values{"MoveType": {moveType}}
		ctx.Request = httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(values.Encode()))
		ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return ctx
	}
	move, err := s.getMoveFromForm(request("Current Player Hit"), blackjackGame)
	if err != nil {
		t.Fatalf("zero-input Current Player Hit should preserve TargetPlayerIndex default: %v", err)
	}
	target, err := move.Reader().PlayerIndexProp("TargetPlayerIndex")
	if err != nil {
		t.Fatalf("reading TargetPlayerIndex: %v", err)
	}
	if target != blackjackGame.CurrentState().CurrentPlayerIndex() {
		t.Errorf("TargetPlayerIndex = %d, want current player %d", target, blackjackGame.CurrentState().CurrentPlayerIndex())
	}

	tictactoeManager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("building tictactoe manager: %v", err)
	}
	tictactoeGame, err := tictactoeManager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building tictactoe game: %v", err)
	}
	if _, err := s.getMoveFromForm(request("Place Token"), tictactoeGame); err == nil || !strings.Contains(err.Error(), "Slot") {
		t.Fatalf("missing required Slot should fail loudly, got %v", err)
	}
}

// TestLegalMoveFormPreviewMatchesMoveLegalAndDoesNotApply pins the two core
// promises of the movePreview endpoint's legality builder (legalMoveForm):
// (1) it reports exactly the authoritative move.Legal() verdict + error the
// real ProposeMove gate would, and (2) it is side-effect-free — previewing must
// never advance the game, so a client can call it on every keystroke.
func TestLegalMoveFormPreviewMatchesMoveLegalAndDoesNotApply(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	move := game.MoveByName("Opted In")
	if move == nil {
		t.Fatal("could not find the Opted In move")
	}
	state := game.CurrentState()
	versionBefore := game.Version()

	form := s.legalMoveForm(game, state, move, 0)

	// (1) legality parity with the authoritative move.Legal (the ground truth
	// game.ProposeMove itself gates on).
	legalErr := move.Legal(state, 0)
	wantLegal := legalErr == nil
	if form.LegalForPlayer != wantLegal {
		t.Errorf("preview LegalForPlayer = %v, want %v (move.Legal(state, 0) == nil)", form.LegalForPlayer, wantLegal)
	}
	if !wantLegal {
		if form.LegalForPlayerError != legalErr.Error() {
			t.Errorf("preview LegalForPlayerError = %q, want %q (verbatim move.Legal error)", form.LegalForPlayerError, legalErr.Error())
		}
	}

	// LegalForAnyone parity with move.Legal(state, AdminPlayerIndex).
	wantAnyone := move.Legal(state, boardgame.AdminPlayerIndex) == nil
	if form.LegalForAnyone != wantAnyone {
		t.Errorf("preview LegalForAnyone = %v, want %v", form.LegalForAnyone, wantAnyone)
	}

	// (2) side-effect-free: previewing must not advance the game version.
	if got := game.Version(); got != versionBefore {
		t.Errorf("preview advanced the game version %d -> %d; the preview path must never apply a move", versionBefore, got)
	}
}

func TestDoMakeMoveReturnsStructuredStaleSnapshot(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	move := game.MoveByName("Opted In")
	if move == nil {
		t.Fatal("could not find the Opted In move")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/game/test/id/move", nil)
	s := &Server{logger: logrus.New()}
	r := s.newRenderer(c)
	versionBefore := game.Version()
	staleVersion := game.Version() - 1
	s.doMakeMove(r, game, boardgame.AdminPlayerIndex, move, &staleVersion)
	if game.Version() != versionBefore {
		t.Fatalf("stale proposal mutated game version from %d to %d", versionBefore, game.Version())
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["Code"] != "STALE_SNAPSHOT" {
		t.Fatalf("Code = %v; want STALE_SNAPSHOT; response: %s", response["Code"], w.Body.String())
	}
	if int(response["ExpectedVersion"].(float64)) != staleVersion ||
		int(response["ActualVersion"].(float64)) != game.Version() {
		t.Fatalf("version metadata = %v; game version = %d", response, game.Version())
	}
}

func TestMovePreviewExpectedVersionIsStructuredAndSideEffectFree(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	version := game.Version()
	request := func(expected string) map[string]any {
		t.Helper()
		body := url.Values{
			"MoveType": {"Opted In"}, "ExpectedVersion": {expected}, "TargetPlayerIndex": {"0"},
		}.Encode()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/game/test/id/movePreview", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		s := &Server{logger: logrus.New()}
		s.setGame(c, game)
		s.setViewingAsPlayer(c, boardgame.AdminPlayerIndex)
		s.movePreviewHandler(c)
		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response %q: %v", w.Body.String(), err)
		}
		return response
	}

	matching := request(strconv.Itoa(version))
	if matching["Status"] != "Success" || matching["Form"] == nil {
		t.Fatalf("matching preview = %v; want Success with Form", matching)
	}
	if game.Version() != version {
		t.Fatalf("matching preview mutated version from %d to %d", version, game.Version())
	}

	staleVersion := version + 1
	stale := request(strconv.Itoa(staleVersion))
	if stale["Code"] != "STALE_SNAPSHOT" || stale["Form"] != nil {
		t.Fatalf("stale preview = %v; want structured stale without Form", stale)
	}
	if int(stale["ExpectedVersion"].(float64)) != staleVersion ||
		int(stale["ActualVersion"].(float64)) != version {
		t.Fatalf("stale metadata = %v; want %d -> %d", stale, staleVersion, version)
	}
	for _, invalid := range []string{"-1", "not-a-version"} {
		response := request(invalid)
		if response["Status"] != "Failure" || response["Form"] != nil {
			t.Fatalf("ExpectedVersion %q response = %v; want Failure without Form", invalid, response)
		}
	}
}

func TestMovePreviewBatchExpectedVersionIsStructuredAndSideEffectFree(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	version := game.Version()
	request := func(expected int) map[string]any {
		t.Helper()
		body, err := json.Marshal(map[string]any{
			"MoveType":        "Opted In",
			"ExpectedVersion": expected,
			"Candidates":      []map[string]any{{"ID": "candidate:[0]", "Args": map[string]string{"TargetPlayerIndex": "0"}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/game/test/id/movePreviewBatch", strings.NewReader(string(body)))
		c.Request.Header.Set("Content-Type", "application/json")
		s := &Server{logger: logrus.New()}
		s.setGame(c, game)
		s.setViewingAsPlayer(c, boardgame.AdminPlayerIndex)
		s.movePreviewBatchHandler(c)
		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response %q: %v", w.Body.String(), err)
		}
		return response
	}

	matching := request(version)
	if matching["Status"] != "Success" || matching["Results"] == nil {
		t.Fatalf("matching batch preview = %v; want Success with Results", matching)
	}
	matchingResults := matching["Results"].([]any)
	if got := matchingResults[0].(map[string]any)["ID"]; got != "candidate:[0]" {
		t.Fatalf("batch preview correlation ID = %v; want candidate:[0]", got)
	}
	if game.Version() != version {
		t.Fatalf("matching batch preview mutated version from %d to %d", version, game.Version())
	}

	staleVersion := version + 1
	stale := request(staleVersion)
	if stale["Code"] != "STALE_SNAPSHOT" || stale["Results"] != nil {
		t.Fatalf("stale batch preview = %v; want structured stale without Results", stale)
	}
	if int(stale["ExpectedVersion"].(float64)) != staleVersion ||
		int(stale["ActualVersion"].(float64)) != version {
		t.Fatalf("stale metadata = %v; want %d -> %d", stale, staleVersion, version)
	}
}

// TestLegalMoveFormsBatch pins the batch preview core (legalMoveFormsBatch),
// which computes legality for many candidate arg-sets of ONE move type against
// a single state in one call — the primitive that lets a client gray a whole
// board's candidate targets in a single round-trip. The promises: each result
// matches the authoritative move.Legal() for that candidate's arg-bound move,
// results stay in candidate order (so the client can correlate by index), a
// malformed candidate soft-fails to Legal:false without killing the batch or
// leaking into its neighbors, an invalid move type is a whole-batch error, and
// the whole thing never advances the game (safe on every keystroke).
func TestLegalMoveFormsBatch(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}
	state := game.CurrentState()
	proposer := boardgame.PlayerIndex(0)

	// expectedFor is the independent oracle: bind a fresh move the same way the
	// batch does, then ask the authoritative move.Legal() (the exact call
	// ProposeMove gates on) for the verdict + verbatim error.
	expectedFor := func(t *testing.T, moveType string, args map[string]string) (bool, string) {
		t.Helper()
		m := game.MoveByName(moveType)
		if m == nil {
			t.Fatalf("could not find move %q", moveType)
		}
		if err := bindMoveFields(m, func(n string) (string, bool) { v, ok := args[n]; return v, ok }); err != nil {
			t.Fatalf("independent bind of %q failed: %v", moveType, err)
		}
		legalErr := m.Legal(state, proposer)
		if legalErr != nil {
			return false, legalErr.Error()
		}
		return true, ""
	}

	t.Run("parity and order across varying args", func(t *testing.T) {
		// "Opted In" embeds moves.CurrentPlayer, so it has a TargetPlayerIndex
		// field to vary per candidate.
		candidates := []movePreviewBatchCandidate{
			{Args: map[string]string{"TargetPlayerIndex": "0"}},
			{Args: map[string]string{"TargetPlayerIndex": "1"}},
			{Args: map[string]string{"TargetPlayerIndex": "0"}},
		}
		versionBefore := game.Version()

		results, err := s.legalMoveFormsBatch(game, state, "Opted In", candidates, proposer)
		if err != nil {
			t.Fatalf("unexpected whole-batch error: %v", err)
		}
		if len(results) != len(candidates) {
			t.Fatalf("got %d results, want %d (exactly one per candidate, in order)", len(results), len(candidates))
		}
		for i, cand := range candidates {
			wantLegal, wantErr := expectedFor(t, "Opted In", cand.Args)
			if results[i].Legal != wantLegal {
				t.Errorf("candidate %d: Legal = %v, want %v (move.Legal parity)", i, results[i].Legal, wantLegal)
			}
			if results[i].Error != wantErr {
				t.Errorf("candidate %d: Error = %q, want %q (verbatim move.Legal error)", i, results[i].Error, wantErr)
			}
		}

		// side-effect-free.
		if got := game.Version(); got != versionBefore {
			t.Errorf("batch preview advanced game version %d -> %d; must never apply", versionBefore, got)
		}
	})

	t.Run("legal candidate (fieldless Opaque move)", func(t *testing.T) {
		// "Opaque" is moves.Default with no fields and no failing precondition
		// — it exercises the Legal:true branch.
		results, err := s.legalMoveFormsBatch(game, state, "Opaque", []movePreviewBatchCandidate{{Args: map[string]string{}}}, proposer)
		if err != nil {
			t.Fatalf("unexpected whole-batch error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		wantLegal, wantErr := expectedFor(t, "Opaque", map[string]string{})
		if !wantLegal {
			t.Fatalf("test precondition: expected Opaque to be legal for proposer 0, but move.Legal said %q", wantErr)
		}
		if !results[0].Legal || results[0].Error != "" {
			t.Errorf("Opaque result = {Legal:%v Error:%q}, want {Legal:true Error:\"\"}", results[0].Legal, results[0].Error)
		}
	})

	t.Run("malformed candidate soft-fails without killing the batch", func(t *testing.T) {
		candidates := []movePreviewBatchCandidate{
			{Args: map[string]string{"TargetPlayerIndex": "0"}},          // fine
			{Args: map[string]string{"TargetPlayerIndex": "not-an-int"}}, // unbindable
			{Args: map[string]string{"TargetPlayerIndex": "1"}},          // fine, must still be evaluated
		}
		results, err := s.legalMoveFormsBatch(game, state, "Opted In", candidates, proposer)
		if err != nil {
			t.Fatalf("a malformed candidate must not fail the whole batch, got error: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("got %d results, want 3 (malformed candidate still occupies its slot)", len(results))
		}
		if results[1].Legal || results[1].Error == "" {
			t.Errorf("malformed candidate result = {Legal:%v Error:%q}, want Legal:false with a non-empty bind error", results[1].Legal, results[1].Error)
		}
		// Neighbors match their own oracle (the bad candidate didn't leak).
		for _, i := range []int{0, 2} {
			wantLegal, wantErr := expectedFor(t, "Opted In", candidates[i].Args)
			if results[i].Legal != wantLegal || results[i].Error != wantErr {
				t.Errorf("candidate %d contaminated by malformed neighbor: got {Legal:%v Error:%q}, want {Legal:%v Error:%q}", i, results[i].Legal, results[i].Error, wantLegal, wantErr)
			}
		}
	})

	t.Run("invalid move type is a whole-batch error", func(t *testing.T) {
		results, err := s.legalMoveFormsBatch(game, state, "No Such Move", []movePreviewBatchCandidate{{Args: map[string]string{}}}, proposer)
		if err == nil {
			t.Fatalf("expected a whole-batch error for an invalid move type, got results %#v", results)
		}
	})

	t.Run("a batch over the candidate cap is rejected whole (DoS guard)", func(t *testing.T) {
		over := make([]movePreviewBatchCandidate, maxLegalPreviewBatchCandidates+1)
		for i := range over {
			over[i] = movePreviewBatchCandidate{Args: map[string]string{"TargetPlayerIndex": "0"}}
		}
		results, err := s.legalMoveFormsBatch(game, state, "Opted In", over, proposer)
		if err == nil {
			t.Fatalf("expected a whole-batch error for %d candidates (> cap %d), got %d results", len(over), maxLegalPreviewBatchCandidates, len(results))
		}
		// At the cap exactly, it must still work (the bound is inclusive).
		atCap := make([]movePreviewBatchCandidate, maxLegalPreviewBatchCandidates)
		for i := range atCap {
			atCap[i] = movePreviewBatchCandidate{Args: map[string]string{"TargetPlayerIndex": "0"}}
		}
		if _, err := s.legalMoveFormsBatch(game, state, "Opted In", atCap, proposer); err != nil {
			t.Fatalf("a batch exactly at the cap (%d) must be allowed, got error: %v", maxLegalPreviewBatchCandidates, err)
		}
	})
}

// TestMovePreviewBatchHandler exercises the gin HTTP handler itself (not just the
// legalMoveFormsBatch core): the POST-only method guard, JSON body parsing +
// its error envelope, the candidate-cap / body-size DoS guards, and that a valid
// request renders an ordered Results array. getGame/effectivePlayerIndex read
// from the gin context, so the fixture game + viewing player are injected
// directly (no full managers/storage wiring needed).
func TestMovePreviewBatchHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	s := &Server{logger: logrus.New()}

	// call invokes the handler with the given method + raw body and returns the
	// HTTP status and the decoded response envelope.
	call := func(method, body string) (int, map[string]interface{}) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(method, "/api/game/x/y/movePreviewBatch", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxGameKey, game)
		c.Set(ctxViewingPlayerAsKey, boardgame.PlayerIndex(0))
		s.movePreviewBatchHandler(c)
		var parsed map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("response was not JSON (%q): %v", w.Body.String(), err)
		}
		return w.Code, parsed
	}

	t.Run("valid batch renders an ordered Results array", func(t *testing.T) {
		_, resp := call(http.MethodPost, `{"MoveType":"Opted In","Candidates":[{"Args":{"TargetPlayerIndex":"0"}},{"Args":{"TargetPlayerIndex":"1"}}]}`)
		if resp["Status"] != "Success" {
			t.Fatalf("want Success, got Status=%v Error=%v", resp["Status"], resp["Error"])
		}
		results, ok := resp["Results"].([]interface{})
		if !ok || len(results) != 2 {
			t.Fatalf("want 2 Results, got %#v", resp["Results"])
		}
	})

	t.Run("correlation IDs are all-or-none, unique, and bounded", func(t *testing.T) {
		bodies := []string{
			`{"MoveType":"Opted In","Candidates":[{"ID":"same","Args":{"TargetPlayerIndex":"0"}},{"ID":"same","Args":{"TargetPlayerIndex":"1"}}]}`,
			`{"MoveType":"Opted In","Candidates":[{"ID":"one","Args":{"TargetPlayerIndex":"0"}},{"Args":{"TargetPlayerIndex":"1"}}]}`,
			`{"MoveType":"Opted In","Candidates":[{"ID":"` + strings.Repeat("x", maxLegalPreviewCandidateIDBytes+1) + `","Args":{"TargetPlayerIndex":"0"}}]}`,
		}
		for _, body := range bodies {
			_, resp := call(http.MethodPost, body)
			if resp["Status"] != "Failure" {
				t.Fatalf("invalid correlation IDs should render Failure: %v", resp)
			}
		}
	})

	t.Run("malformed JSON renders the Failure envelope, not a panic", func(t *testing.T) {
		_, resp := call(http.MethodPost, `{not valid json`)
		if resp["Status"] != "Failure" {
			t.Errorf("malformed JSON should render Failure, got %v", resp["Status"])
		}
	})

	t.Run("GET is rejected (POST-only)", func(t *testing.T) {
		_, resp := call(http.MethodGet, `{"MoveType":"Opted In","Candidates":[]}`)
		if resp["Status"] != "Failure" {
			t.Errorf("GET should be rejected, got %v", resp["Status"])
		}
	})

	t.Run("over-cap candidate count renders Failure (DoS guard)", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString(`{"MoveType":"Opted In","Candidates":[`)
		for i := 0; i <= maxLegalPreviewBatchCandidates; i++ { // one over the cap
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`{"Args":{"TargetPlayerIndex":"0"}}`)
		}
		sb.WriteString(`]}`)
		_, resp := call(http.MethodPost, sb.String())
		if resp["Status"] != "Failure" {
			t.Errorf("a batch over the candidate cap should render Failure, got %v", resp["Status"])
		}
	})

	t.Run("oversized body renders Failure (MaxBytesReader guard)", func(t *testing.T) {
		// A body past the 1 MiB cap must be shed by http.MaxBytesReader so
		// BindJSON errors out rather than buffering it all.
		big := `{"MoveType":"` + strings.Repeat("x", maxLegalPreviewBodyBytes+1024) + `","Candidates":[]}`
		_, resp := call(http.MethodPost, big)
		if resp["Status"] != "Failure" {
			t.Errorf("an oversized body should render Failure, got %v", resp["Status"])
		}
	})
}

// TestMovePreviewBatchHandlerAdminPerspective pins must-fix #2 at the HTTP layer:
// an admin previewing "as player N" (admin=1&player=N, with admin allowed) must
// have legality evaluated as player N — not as the observer/session player. A
// regression that dropped the player/admin resolution would evaluate every
// candidate as the current player and wrongly report legal, re-graying the whole
// board for the admin. Observable via a real tictactoe game: only the current
// player may place on the empty board.
func TestMovePreviewBatchHandlerAdminPerspective(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("building tictactoe manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building tictactoe game: %v", err)
	}
	s := &Server{logger: logrus.New()}

	const body = `{"MoveType":"Place Token","Candidates":[{"Args":{"Slot":"0"}},{"Args":{"Slot":"1"}},{"Args":{"Slot":"2"}}]}`

	// callAsAdmin previews as an admin acting as the given player: admin=1 &
	// player=N in the query, ctxAdminAllowedKey set so calcIsAdmin honors it.
	callAsAdmin := func(t *testing.T, player int) []interface{} {
		t.Helper()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/game/tictactoe/g/movePreviewBatch?admin=1&player="+strconv.Itoa(player), strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxGameKey, game)
		c.Set(ctxAdminAllowedKey, true)
		s.movePreviewBatchHandler(c)
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("response not JSON (%q): %v", w.Body.String(), err)
		}
		if resp["Status"] != "Success" {
			t.Fatalf("admin as player %d: want Success, got Status=%v Error=%v", player, resp["Status"], resp["Error"])
		}
		return resp["Results"].([]interface{})
	}

	// Current player (0): the empty board is fully open.
	for i, r := range callAsAdmin(t, 0) {
		if m := r.(map[string]interface{}); m["Legal"] != true {
			t.Errorf("admin as current player 0, slot %d: Legal=%v, want true", i, m["Legal"])
		}
	}
	// Non-current player (1): not their turn -> every slot illegal. The proposer
	// param must be honored; ignoring it (evaluating as current player 0) would
	// wrongly return legal here — the exact regression must-fix #2 guards.
	for i, r := range callAsAdmin(t, 1) {
		if m := r.(map[string]interface{}); m["Legal"] == true {
			t.Errorf("admin as non-current player 1, slot %d: Legal=true, want false (not their turn)", i)
		}
	}
}

// TestLegalMoveFormsBatchRealTictactoeGame runs the batch preview against a REAL
// example game (not the synthetic fixture) to prove it composes with real move
// legality end-to-end: the tictactoe "Place Token" move embeds moves.CurrentPlayer
// (so its TargetPlayerIndex field is set by DefaultsForState) and gates on
// LegalCustom (token availability + MayMoveToSlot). Candidates supply only the
// varying Slot and rely on TargetPlayerIndex's default — exactly what the client
// board (previewSpec) does. Also pins that batch legality equals move.Legal and
// that previewing never advances the game.
func TestLegalMoveFormsBatchRealTictactoeGame(t *testing.T) {
	manager, err := boardgame.NewGameManager(tictactoe.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("building tictactoe manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building tictactoe game: %v", err)
	}

	s := &Server{}
	state := game.CurrentState()
	const moveType = "Place Token"
	if game.MoveByName(moveType) == nil {
		t.Fatalf("tictactoe has no %q move (auto-name changed?)", moveType)
	}
	current := state.CurrentPlayerIndex()
	versionBefore := game.Version()

	// One candidate per board slot, supplying ONLY the varying Slot — the batch
	// keeps TargetPlayerIndex at its DefaultsForState value (the current player).
	candidates := make([]movePreviewBatchCandidate, 9)
	for i := range candidates {
		candidates[i] = movePreviewBatchCandidate{Args: map[string]string{"Slot": strconv.Itoa(i)}}
	}

	// As the current player, the empty board is fully open — every slot legal.
	curResults, err := s.legalMoveFormsBatch(game, state, moveType, candidates, current)
	if err != nil {
		t.Fatalf("batch as current player: %v", err)
	}
	for i, r := range curResults {
		if !r.Legal {
			t.Errorf("current player, slot %d: Legal=false (%q); an empty board should be fully open (default TargetPlayerIndex must fill in)", i, r.Error)
		}
	}

	// Parity: each batch verdict equals the authoritative move.Legal for a move
	// bound the same way (Slot only, default TargetPlayerIndex).
	for i := range candidates {
		m := game.MoveByName(moveType)
		slot := candidates[i].Args["Slot"]
		if err := bindMoveFields(m, func(name string) (string, bool) {
			if name == "Slot" {
				return slot, true
			}
			return "", false
		}); err != nil {
			t.Fatalf("independent bind of slot %d: %v", i, err)
		}
		if want := m.Legal(state, current) == nil; curResults[i].Legal != want {
			t.Errorf("slot %d: batch Legal=%v, want %v (move.Legal parity)", i, curResults[i].Legal, want)
		}
	}

	// As a non-current player it isn't their turn, so the whole board grays.
	other := boardgame.PlayerIndex(1)
	if other == current {
		other = boardgame.PlayerIndex(0)
	}
	otherResults, err := s.legalMoveFormsBatch(game, state, moveType, candidates, other)
	if err != nil {
		t.Fatalf("batch as non-current player: %v", err)
	}
	for i, r := range otherResults {
		if r.Legal {
			t.Errorf("non-current player %d, slot %d: Legal=true; expected illegal (not their turn)", other, i)
		}
	}

	if got := game.Version(); got != versionBefore {
		t.Errorf("batch preview advanced a real game's version %d -> %d; must never apply", versionBefore, got)
	}
}
