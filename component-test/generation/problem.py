from typing import Any

import unified_planning.shortcuts as up

up.get_environment().credits_stream = None

####################
# User Types
####################
Player = up.UserType("Player")
Region = up.UserType("Region")
Continent = up.UserType("Continent")
Phase = up.UserType("Phase")

####################
# Fluents
####################

Adjacent = up.Fluent("Adjacent", up.BoolType(), region1=Region, region2=Region)
DeployableTroops = up.Fluent("DeployableTroops", up.IntType(), player=Player)

# Region related fluents
Owns = up.Fluent("Owns", up.BoolType(), player=Player, region=Region)
TroopsOn = up.Fluent("TroopsOn", up.IntType(), region=Region)

# Phase/turn related fluents
Turn = up.Fluent("Turn", up.BoolType(), player=Player)
NextPlayer = up.Fluent("NextPlayer", up.BoolType(), current=Player, next=Player)
CurrentPhase = up.Fluent("CurrentPhase", up.BoolType(), phase=Phase)

# Conquer-related fluents
HasWonAttack = up.Fluent("hasWonAttack", up.BoolType(), from_region=Region, to_region=Region)

####################
# Objects
####################
deploy_phase = up.Object("deploy", Phase)
attack_phase = up.Object("attack", Phase)
conquer_phase = up.Object("conquer", Phase)


####################
# Actions
####################

def check_turn_and_phase(player: Player, phase: Phase) -> up.BoolType():
    return Turn(player) & CurrentPhase(phase)


def pass_to_next_phase(current_phase: Phase, next_phase: Phase) -> list[tuple[up.Fluent, Any]]:
    """Return the effect to pass from current_phase to next_phase."""
    return [(CurrentPhase(current_phase), False), (CurrentPhase(next_phase), True)]


# simplification: we deploy all troops at once
deploy_action = up.InstantaneousAction("deploy", player=Player, region=Region)
deploy_action.add_precondition(
    check_turn_and_phase(deploy_action.player, deploy_phase) &
    Owns(deploy_action.player, deploy_action.region) &
    (DeployableTroops(deploy_action.player) >= 0)
)
deploy_action.add_effect(DeployableTroops(deploy_action.player),
                         value=0)
deploy_action.add_increase_effect(TroopsOn(deploy_action.region),
                                  value=DeployableTroops(deploy_action.player))
deploy_action.add_effect(CurrentPhase(deploy_phase), value=False)
deploy_action.add_effect(CurrentPhase(attack_phase), value=True)

# simplifications:
# - attack until conquering.
# - we always win the attack
attack_until_conquering_action = up.InstantaneousAction("attack", player=Player, source=Region, target=Region)
attack_until_conquering_action.add_precondition(
    check_turn_and_phase(attack_until_conquering_action.player, attack_phase) &
    Adjacent(attack_until_conquering_action.source, attack_until_conquering_action.target) &
    Owns(attack_until_conquering_action.player, attack_until_conquering_action.source) &
    ~Owns(attack_until_conquering_action.player, attack_until_conquering_action.target) &
    (TroopsOn(attack_until_conquering_action.source) >= 1)
)
attack_until_conquering_action.add_effect(
    TroopsOn(attack_until_conquering_action.target),
    value=0,
)
attack_until_conquering_action.add_effect(
    HasWonAttack(attack_until_conquering_action.source, attack_until_conquering_action.target),
    value=True,
)
attack_until_conquering_action.add_effect(
    CurrentPhase(attack_phase),
    value=False,
)
attack_until_conquering_action.add_effect(
    CurrentPhase(conquer_phase),
    value=True,
)

# simplification: we move all but one troop
conquer_action = up.InstantaneousAction("conquer", conquering_player=Player, conquered_player=Player, source=Region,
                                        target=Region)
conquer_action.add_precondition(
    check_turn_and_phase(conquer_action.conquering_player, conquer_phase) &
    HasWonAttack(conquer_action.source, conquer_action.target) &
    Owns(conquer_action.conquering_player, conquer_action.source) &
    Owns(conquer_action.conquered_player, conquer_action.target)
)
conquer_action.add_effect(
    Owns(conquer_action.conquering_player, conquer_action.target),
    value=True,
)
conquer_action.add_effect(
    Owns(conquer_action.conquered_player, conquer_action.target),
    value=False,
)
conquer_action.add_effect(
    TroopsOn(conquer_action.source),
    value=1,
)
conquer_action.add_increase_effect(
    TroopsOn(conquer_action.target),
    value=TroopsOn(conquer_action.source) - 1,
)
conquer_action.add_effect(
    HasWonAttack(conquer_action.source, conquer_action.target),
    value=False,
)
conquer_action.add_effect(
    CurrentPhase(conquer_phase),
    value=False,
)
conquer_action.add_effect(
    CurrentPhase(attack_phase),
    value=True,
)
