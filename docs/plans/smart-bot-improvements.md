# Smart Bot Improvements

Deferred from the performance & scalability work. The Phase 1 smart bot (random valid moves) is sufficient for load testing. These phases improve game realism and strategic behavior.

## Phase 2: Board Evaluator + Move Scoring

Replace random move selection with weighted scoring.

**Board evaluator factors:**
- Territory count (more = better)
- Continent control (full continents = bonus armies)
- Border exposure (fewer borders to defend = safer)
- Army concentration (avoid spreading too thin)
- Card set readiness (close to trading = higher value)

**Move scoring:**
- Deploy: weight toward border regions and continent-completion targets
- Attack: score by expected outcome (army advantage), strategic value (continent completion), risk tolerance
- Reinforce: strengthen weakest borders, consolidate toward contested continents
- Conquer: move enough troops to hold, but don't leave source exposed

**Files to create:**
- `perf-test/internal/player/smart/evaluator.go` — board state scoring
- `perf-test/internal/player/smart/scorer.go` — move scoring and weighted selection

## Phase 3: Attack Probability Model

Model Risk combat dice probabilities for better attack/retreat decisions.

**Combat tables:**
- Pre-computed win probability tables for attacker vs defender army counts
- Expected army loss tables for attacker and defender
- Threshold-based attack decisions: only attack when expected outcome is favorable

**Files to create:**
- `perf-test/internal/player/smart/combat.go` — probability tables and expected values
- Update `strategy.go` attack logic to use combat model

## Phase 4: Personality System

Per-player personality traits, think time variation, and CLI controls.

**Personality traits:**
- Aggression (0-1): likelihood to attack marginal targets
- Risk tolerance (0-1): willingness to attack with small advantages
- Expansion preference (0-1): prioritize territory count vs consolidation
- Card hoarding (0-1): preference to hold cards vs trade early

**CLI flags:**
- `--strategy=mixed` — each player gets a random personality
- `--strategy-seed=N` — reproducible personality assignment
- `--think-time-variance=500ms` — random variation around base think time per personality

**Files to create:**
- `perf-test/internal/player/smart/personality.go` — already exists (Phase 1 stub), expand with full trait system
- Update `strategy.go` to use personality traits in scoring
