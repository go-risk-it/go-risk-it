# Game Rules

GO Risk-It implements a variant of the classic [Risk](https://en.wikipedia.org/wiki/Risk_(game)) board game with secret missions.

## Objective

Each player is assigned a **secret mission** at the start of the game. The first player to complete their mission wins. Unlike classic Risk, the goal is not world domination — it's fulfilling your specific objective.

## Mission Types

| Mission | Condition |
|---------|-----------|
| **Conquer 18 territories** | Own 18 territories, each defended by at least 2 armies |
| **Conquer 24 territories** | Own 24 territories (any troop count) |
| **Two continents** | Conquer two specific continents (e.g., North America + Africa) |
| **Two continents + one** | Conquer two specific continents + any third continent of your choice |
| **Eliminate player** | Eliminate a specific player. If another player eliminates them first, conquer 24 territories instead |

### Continental Missions

The following continent pairs can be assigned:

- North America + Africa
- North America + Oceania
- Asia + South America
- Asia + Africa
- Europe + South America + one of your choice
- Europe + Oceania + one of your choice

## Turn Phases

Each turn consists of five phases, executed in order:

### 1. Cards

Trade sets of three cards for bonus troops. Card combinations:

- 3 of the same type (3 Infantry, 3 Cavalry, or 3 Artillery)
- 1 of each type (1 Infantry + 1 Cavalry + 1 Artillery)
- Any 2 cards + 1 Jolly (wild card)

**Rules:**
- Skip if you have fewer than 3 cards
- **Must** trade if you have 5 or more cards

### 2. Deploy

Place your available troops on regions you own. Troop count is calculated from:

- Number of territories owned (territories / 3, minimum 3)
- Continent bonuses (own all territories in a continent)
- Card trade-in bonuses

### 3. Attack

Attack an adjacent enemy region from one of your regions. You must have at least 2 troops in the attacking region (one must stay behind). Dice determine the outcome:

- Attacker rolls up to 3 dice (one fewer than troops committed)
- Defender rolls up to 3 dice (one per defending troop, max 3)
- Highest dice are compared in order — ties go to the defender

You can attack as many times as you want, or skip this phase entirely.

### 4. Conquer

After a successful attack, move troops from the attacking region into the conquered region. You must move at least as many troops as you used to attack, and at least one troop must remain in the source region.

### 5. Reinforce

Move troops between your own regions that are connected through a chain of your territories. This is optional — you can end your turn without reinforcing.

## Simultaneous Victory

It is possible for two players to win simultaneously. For example:

1. Player A has the mission "Eliminate Player C"
2. Player A currently owns 24 territories
3. Player B has the mission "Conquer North America + Oceania"
4. Player B controls all of North America and is missing one region in Oceania
5. Player C controls that last region in Oceania
6. Player B eliminates Player C → Player B completes their mission
7. Player A's mission changes to "Conquer 24 territories" → already satisfied

Both Player A and Player B win at the same time.
