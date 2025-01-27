import unified_planning.shortcuts as up

from fluents import Adjacent, DeployableTroops, Owns, TroopsOn, BonusTroops, Turn, \
    NextPlayer, CurrentPhase, HasWonAttack
from shortcuts import check_turn_and_phase, owns_continent
from user_types import Player, Region, Continent, Phase

up.get_environment().credits_stream = None

####################
# Objects
####################
phase_deploy = up.Object("deploy", Phase)
phase_attack = up.Object("attack", Phase)
phase_conquer = up.Object("conquer", Phase)
phase_reinforce = up.Object("reinforce", Phase)
phase_cards = up.Object("cards", Phase)

####################
# Actions
####################

# Deploy action
# simplification: we deploy all troops at once
action_deploy = up.InstantaneousAction("deploy", player=Player, region=Region)
action_deploy.add_precondition(
    check_turn_and_phase(action_deploy.player, phase_deploy) &
    Owns(action_deploy.player, action_deploy.region) &
    (DeployableTroops(action_deploy.player) >= 0)
)
action_deploy.add_effect(DeployableTroops(action_deploy.player),
                         value=0)
action_deploy.add_increase_effect(TroopsOn(action_deploy.region),
                                  value=DeployableTroops(action_deploy.player))
action_deploy.add_effect(CurrentPhase(phase_deploy), value=False)
action_deploy.add_effect(CurrentPhase(phase_attack), value=True)

# Attack action
# simplifications:
# - attack until conquering.
# - we always win the attack
action_attack_until_conquering = up.InstantaneousAction("attack_until_conquering", player=Player, source=Region,
                                                        target=Region)
action_attack_until_conquering.add_precondition(
    check_turn_and_phase(action_attack_until_conquering.player, phase_attack) &
    Adjacent(action_attack_until_conquering.source, action_attack_until_conquering.target) &
    Owns(action_attack_until_conquering.player, action_attack_until_conquering.source) &
    ~Owns(action_attack_until_conquering.player, action_attack_until_conquering.target) &
    (TroopsOn(action_attack_until_conquering.source) >= 1)
)
action_attack_until_conquering.add_effect(
    TroopsOn(action_attack_until_conquering.target),
    value=0,
)
action_attack_until_conquering.add_effect(
    HasWonAttack(action_attack_until_conquering.source, action_attack_until_conquering.target),
    value=True,
)
action_attack_until_conquering.add_effect(
    CurrentPhase(phase_attack),
    value=False,
)
action_attack_until_conquering.add_effect(
    CurrentPhase(phase_conquer),
    value=True,
)

# Conquer action
# simplification: we move all but one troop
action_conquer = up.InstantaneousAction("conquer", conquering_player=Player, conquered_player=Player, source=Region,
                                        target=Region)
action_conquer.add_precondition(
    check_turn_and_phase(action_conquer.conquering_player, phase_conquer) &
    HasWonAttack(action_conquer.source, action_conquer.target) &
    Owns(action_conquer.conquering_player, action_conquer.source) &
    Owns(action_conquer.conquered_player, action_conquer.target)
)
action_conquer.add_effect(
    Owns(action_conquer.conquering_player, action_conquer.target),
    value=True,
)
action_conquer.add_effect(
    Owns(action_conquer.conquered_player, action_conquer.target),
    value=False,
)
action_conquer.add_effect(
    TroopsOn(action_conquer.source),
    value=1,
)
action_conquer.add_increase_effect(
    TroopsOn(action_conquer.target),
    value=TroopsOn(action_conquer.source) - 1,
)
action_conquer.add_effect(
    HasWonAttack(action_conquer.source, action_conquer.target),
    value=False,
)
action_conquer.add_effect(
    CurrentPhase(phase_conquer),
    value=False,
)
action_conquer.add_effect(
    CurrentPhase(phase_attack),
    value=True,
)

# Reinforce action
# simplifications:
# - reinforce only between adjacent regions
# - we move all troops
action_reinforce = up.InstantaneousAction("reinforce", player=Player, source=Region, target=Region, next_player=Player)
action_reinforce.add_precondition(
    check_turn_and_phase(action_reinforce.player, phase_reinforce) &
    Adjacent(action_reinforce.source, action_reinforce.target) &
    Owns(action_reinforce.player, action_reinforce.source) &
    Owns(action_reinforce.player, action_reinforce.target) &
    (TroopsOn(action_reinforce.source) >= 2) &
    NextPlayer(action_reinforce.player, action_reinforce.next_player)
)
action_reinforce.add_effect(
    TroopsOn(action_reinforce.source),
    value=1,
)
action_reinforce.add_increase_effect(
    TroopsOn(action_reinforce.target),
    value=TroopsOn(action_reinforce.source) - 1,
)
action_reinforce.add_effect(
    CurrentPhase(phase_reinforce),
    value=False,
)
action_reinforce.add_effect(
    CurrentPhase(phase_deploy),
    value=True,
)
action_reinforce.add_effect(
    Turn(action_reinforce.player),
    value=False,
)
action_reinforce.add_effect(
    Turn(action_reinforce.next_player),
    value=True,
)

# Advance action
action_advance = up.InstantaneousAction("advance", player=Player, next_player=Player, from_phase=Phase, to_phase=Phase)
action_advance.add_precondition(
    check_turn_and_phase(action_advance.player, action_advance.from_phase) &
    NextPlayer(action_advance.player, action_advance.next_player) &
    up.Or(
        up.Equals(action_advance.from_phase, phase_attack) & up.Equals(action_advance.to_phase, phase_reinforce),
        up.Equals(action_advance.from_phase, phase_reinforce) & up.Equals(action_advance.to_phase, phase_cards),
        up.Equals(action_advance.from_phase, phase_cards) & up.Equals(action_advance.to_phase, phase_deploy),
    )
)
action_advance.add_effect(
    CurrentPhase(action_advance.from_phase),
    value=False,
)
action_advance.add_effect(
    CurrentPhase(action_advance.to_phase),
    value=True,
)
action_advance.add_effect(
    Turn(action_advance.player),
    value=False,
    condition=up.Equals(action_advance.from_phase, phase_reinforce)
)
action_advance.add_effect(
    Turn(action_advance.next_player),
    value=True,
    condition=up.Equals(action_advance.from_phase, phase_reinforce)
)

# When advancing to deploy phase, compute the number of deployable troops
# get 3 troops for each region owned
r = up.Variable("r", Region)
action_advance.add_increase_effect(
    DeployableTroops(action_advance.player),
    value=3,
    condition=up.Equals(action_advance.to_phase, phase_deploy) & Owns(action_advance.player, r),
    forall=[r]
)
# get bonus troops for each continent owned
c = up.Variable("c", Continent)
action_advance.add_increase_effect(
    DeployableTroops(action_advance.player),
    value=BonusTroops(c),
    condition=up.Equals(action_advance.to_phase, phase_deploy) &
              owns_continent(action_advance.player, c),
    forall=[c]
)
