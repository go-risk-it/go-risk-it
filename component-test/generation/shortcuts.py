from unified_planning import shortcuts as ups

from fluents import Owns, BelongsTo, Turn, CurrentPhase
from user_types import Player, Region, Continent, Phase


def check_turn_and_phase(player: Player, phase: Phase) -> ups.BoolType():
    return Turn(player) & CurrentPhase(phase)


def owns_continent(player: Player, continent: Continent) -> ups.BoolType:
    r = ups.Variable("r", Region)
    return ups.Forall(~BelongsTo(r, continent) | Owns(player, r), r)
