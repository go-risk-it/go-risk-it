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

FILE_DIR = os.path.dirname(os.path.abspath(__file__))

from domain import (Region, Player, NextPlayer, Turn, CurrentPhase, HasWonAttack,
                    DeployableTroops, TroopsOn, Owns, Adjacent,
                    phase_deploy, action_deploy, action_attack_until_conquering,
                    action_conquer, action_advance, action_reinforce)
import unified_planning.shortcuts as ups
import unified_planning.engines as upe


# import unified_planning.io as upi


def init_map(map_filename: str, problem: ups.Problem) -> tuple[dict[str, Region], dict[str, list[Region]]]:
    """
    Read a Risk map from a file and initialize the regions and adjacency relations.
    """
    regions = {}
    continents = {}

    with open(map_filename, "r") as map_file:
        map_data = json.load(map_file)

    for layer in map_data["layers"]:
        region_id = layer["id"]
        region = ups.Object(region_id, Region)
        problem.add_object(region)
        regions[region_id] = region
        continents[layer["continent"]] = continents.get(layer["continent"], []) + [region]

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
    pp = list(players.values())
    player_to_regions = {
        players["francesco"]: [],
        players["gabriele"]: [],
        players["giovanni"]: [],
    }
    for i, region in enumerate(regions.values()):
        player = pp[i % len(pp)]
        problem.set_initial_value(Owns(player, region), True)
        player_to_regions[player].append(region)
    for player, regions in player_to_regions.items():
        print(f"Regions of {player}: {sorted(regions, key=lambda r: r.name)}")

    problem.add_fluent(HasWonAttack, default_initial_value=False)

    # actions
    problem.add_action(action_deploy)
    problem.add_action(action_attack_until_conquering)
    problem.add_action(action_conquer)
    problem.add_action(action_reinforce)
    problem.add_action(action_advance)

    return players


def main():
    problem = ups.Problem("Risk")

    regions, continents = init_map(os.path.join(FILE_DIR, "..", "..", "map.json"), problem)
    players = init_game(problem, regions)
    # print(problem)

    problem.add_goal(ups.And([Owns(players["francesco"], region) for region in continents["europe"]]))
    problem.add_goal(ups.And([Owns(players["gabriele"], region) for region in continents["north_america"]]))
    problem.add_goal(ups.And([Owns(players["giovanni"], region) for region in continents["south_america"]]))

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
