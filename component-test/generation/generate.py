"""
Encode Risk game logic to generate games.

Note: the enhsp planner version included in the unified_planning library seems to have a bug.
    I had to download enhsp-20.jar from here https://drive.google.com/file/d/1GfVLQNEgeeNnNeI6HkrCtAUrrSzdSW8g/view?usp=sharing
    and replace the one in the unified_planning library, i.e.,
    venv/lib/python3.12/site-packages/up_enhsp/ENHSP/enhsp.jar
    with the downloaded one.
"""
import json
import os.path

from unified_planning.model import Object

from shortcuts import owns_continent

FILE_DIR = os.path.dirname(os.path.abspath(__file__))

from domain import (phase_deploy, action_deploy, action_attack_until_conquering,
                    action_conquer, action_advance, action_reinforce)
from fluents import (Adjacent, DeployableTroops, Owns, TroopsOn, BelongsTo, BonusTroops, Turn,
                     NextPlayer, CurrentPhase, HasWonAttack)
from user_types import Player, Region, Continent
import unified_planning.shortcuts as ups
import unified_planning.engines as upe
from itertools import cycle


# import unified_planning.io as upi


def init_map(map_filename: str, problem: ups.Problem) -> tuple[dict[str, Object], dict[str, Object]]:
    """
    Read a Risk map from a file and initialize the regions and adjacency relations.
    """
    regions = {}
    continents = {}

    with open(map_filename, "r") as map_file:
        map_data = json.load(map_file)

    problem.add_fluent(BonusTroops, default_initial_value=0)
    for continent_data in map_data["continents"]:
        continent_id: str = continent_data["id"]
        continent = ups.Object(continent_id, Continent)
        problem.add_object(continent)
        continents[continent_id] = continent
        problem.set_initial_value(BonusTroops(continent), continent_data["bonusTroops"])

    problem.add_fluent(BelongsTo, default_initial_value=False)
    for region_data in map_data["layers"]:
        region_id: str = region_data["id"]
        region = ups.Object(region_id, Region)
        problem.add_object(region)
        regions[region_id] = region
        continent = continents[region_data["continent"]]
        problem.set_initial_value(BelongsTo(region, continent), True)

    problem.add_fluent(Adjacent, default_initial_value=False)
    for link in map_data["links"]:
        source_region = regions[link["source"]]
        target_region = regions[link["target"]]
        problem.set_initial_value(Adjacent(source_region, target_region), True)
        problem.set_initial_value(Adjacent(target_region, source_region), True)

    return regions, continents


def init_game(problem: ups.Problem, regions: dict[str, Region]) -> dict[str, Player]:
    """
    initialize the game
    """

    # Players
    players = {name: ups.Object(name, Player) for name in ["francesco", "gabriele", "giovanni"]}
    problem.add_objects(players.values())

    # NextPlayer
    problem.add_fluent(NextPlayer, default_initial_value=False)
    problem.set_initial_value(NextPlayer(players["francesco"], players["gabriele"]), True)
    problem.set_initial_value(NextPlayer(players["gabriele"], players["giovanni"]), True)
    problem.set_initial_value(NextPlayer(players["giovanni"], players["francesco"]), True)

    # Turn
    problem.add_fluent(Turn, default_initial_value=False)
    problem.set_initial_value(Turn(players["francesco"]), True)

    problem.add_fluent(CurrentPhase, default_initial_value=False)
    problem.set_initial_value(CurrentPhase(phase_deploy), True)
    # Initial deployable troops 3
    problem.add_fluent(DeployableTroops, default_initial_value=3)

    # Initial troops on regions
    problem.add_fluent(TroopsOn, default_initial_value=3)

    # Assign regions to players
    problem.add_fluent(Owns, default_initial_value=False)
    player_to_regions = assign_regions_to_player(players, regions)
    for player, regions in player_to_regions.items():
        for region in regions:
            problem.set_initial_value(Owns(player, region), True)

    problem.add_fluent(HasWonAttack, default_initial_value=False)

    # actions
    problem.add_action(action_deploy)
    problem.add_action(action_attack_until_conquering)
    problem.add_action(action_conquer)
    problem.add_action(action_reinforce)
    problem.add_action(action_advance)

    return players


def assign_regions_to_player(players: dict[str, Player], regions: dict[str, Region]) -> dict[Player, list[Region]]:
    player_to_regions = {player: [] for player in players.values()}

    for player, region in zip(cycle(players.values()), regions.values()):
        player_to_regions[player].append(region)

    return player_to_regions


def main():
    problem = ups.Problem("Risk")

    regions, continents = init_map(os.path.join(FILE_DIR, "..", "..", "map.json"), problem)
    players = init_game(problem, regions)
    # print(problem)

    problem.add_goal(owns_continent(players["francesco"], continents["europe"]))
    problem.add_goal(owns_continent(players["gabriele"], continents["north_america"]))
    problem.add_goal(owns_continent(players["giovanni"], continents["asia"]))

    # write pddl file
    # upi.PDDLWriter(problem).write_domain(os.path.join(FILE_DIR, "risk_domain.pddl"))
    # upi.PDDLWriter(problem).write_problem(os.path.join(FILE_DIR, "risk_problem.pddl"))

    with ups.OneshotPlanner(name="enhsp") as planner:
        planner: upe.pddl_planner.PDDLPlanner
        result = planner.solve(problem)
        if result.status == upe.PlanGenerationResultStatus.SOLVED_SATISFICING:
            print(f"Planner returned: {result.plan}")
            print(f"Reached goal: {problem.goals}")
        else:
            print("No plan found: {} {}".format(result.status, result.log_messages))

    # francesco = players["francesco"]
    # gabriele = players["gabriele"]
    # giovanni = players["giovanni"]
    #
    # plan = upp.SequentialPlan([
    #     action_deploy(francesco, regions["alaska"]),
    #     action_attack_until_conquering(francesco, regions["greenland"], regions["ontario"]),
    #     action_conquer(francesco, gabriele, regions["greenland"], regions["ontario"]),
    #     action_advance(francesco, gabriele, phase_attack, phase_reinforce),
    #     action_advance(francesco, gabriele, phase_reinforce, phase_cards),
    #     action_advance(gabriele, giovanni, phase_cards, phase_deploy),
    #     action_deploy(gabriele, regions["quebec"]),
    #     action_attack_until_conquering(gabriele, regions["quebec"], regions["greenland"]),
    #     action_conquer(gabriele, francesco, regions["quebec"], regions["greenland"]),
    #     action_attack_until_conquering(gabriele, regions["greenland"], regions["ontario"]),
    #     action_conquer(gabriele, francesco, regions["greenland"], regions["ontario"]),
    #     action_advance(gabriele, giovanni, phase_attack, phase_reinforce),
    #     action_advance(gabriele, giovanni, phase_reinforce, phase_cards),
    #     action_advance(giovanni, francesco, phase_cards, phase_deploy),
    #     action_deploy(giovanni, regions["ural"]),
    #     action_advance(giovanni, francesco, phase_attack, phase_reinforce),
    #     action_advance(giovanni, francesco, phase_reinforce, phase_cards),
    #     action_advance(francesco, gabriele, phase_cards, phase_deploy),
    #     action_deploy(francesco, regions["afghanistan"]),
    #     action_attack_until_conquering(francesco, regions["afghanistan"], regions["ural"]),
    # ])
    #
    # with ups.SequentialSimulator(problem) as simulator:
    #     simulator: upe.sequential_simulator.UPSequentialSimulator
    #     state = simulator.get_initial_state()
    #     # print(f"Initial state: {state}")
    #
    #     for action in plan.actions:
    #         print(f"Applying action: {action}")
    #         if not simulator.is_applicable(state, action):
    #             print(f"Action not applicable, because of: {simulator.get_unsatisfied_conditions(state, action)}")
    #             break
    #         state: upm.State = simulator.apply_unsafe(state, action)
    #         # print(f"New state: {state}")
    #         # print(f"Remaining battery: {state.get_value(battery)}")
    #     if simulator.is_goal(state):
    #         print("Goal reached!")
    #     else:
    #         print("Goal not reached:") # .format(simulator.get_unsatisfied_goals(state)))


if __name__ == "__main__":
    main()
