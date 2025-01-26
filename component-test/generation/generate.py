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

from problem import Region, Player, NextPlayer, Turn, CurrentPhase, deploy_action, deploy_phase, \
    DeployableTroops, TroopsOn, Owns, attack_until_conquering_action, conquer_action, Adjacent, HasWonAttack
import unified_planning.shortcuts as ups
import unified_planning.engines as upe
import unified_planning.plans as upp
import unified_planning.io as upi


def init_map(map_filename: str, problem: ups.Problem) -> dict[str, Region]:
    """
    Read a Risk map from a file and initialize the regions and adjacency relations.
    """
    regions = {}

    with open(map_filename, "r") as map_file:
        map_data = json.load(map_file)

    for layer in map_data["layers"]:
        region_id = layer["id"]
        region = ups.Object(region_id, Region)
        problem.add_object(region)
        regions[region_id] = region

    problem.add_fluent(Adjacent, default_initial_value=False)

    for link in map_data["links"]:
        source_region = regions[link["source"]]
        target_region = regions[link["target"]]
        problem.set_initial_value(Adjacent(source_region, target_region), True)
        problem.set_initial_value(Adjacent(target_region, source_region), True)

    return regions


def init_game(problem: ups.Problem, regions: dict[str, Region]) -> list[Player]:
    """
    initialize the game
    """

    # Players
    francesco = ups.Object("francesco", Player)
    gabriele = ups.Object("gabriele", Player)
    giovanni = ups.Object("giovanni", Player)
    players = [francesco, gabriele, giovanni]
    problem.add_objects(players)

    # NextPlayer
    problem.add_fluent(NextPlayer, default_initial_value=False)
    problem.set_initial_value(NextPlayer(francesco, gabriele), True)
    problem.set_initial_value(NextPlayer(gabriele, giovanni), True)
    problem.set_initial_value(NextPlayer(giovanni, francesco), True)

    # Turn
    problem.add_fluent(Turn, default_initial_value=False)
    problem.set_initial_value(Turn(francesco), True)

    problem.add_fluent(CurrentPhase, default_initial_value=False)
    problem.set_initial_value(CurrentPhase(deploy_phase), True)
    # Initial deployable troops 3
    problem.add_fluent(DeployableTroops, default_initial_value=3)

    # Initial troops on regions
    problem.add_fluent(TroopsOn, default_initial_value=3)

    # Assign regions to players
    problem.add_fluent(Owns, default_initial_value=False)
    for i, region in enumerate(regions.values()):
        player = players[i % len(players)]
        problem.set_initial_value(Owns(player, region), True)

    problem.add_fluent(HasWonAttack, default_initial_value=False)

    # actions
    problem.add_action(deploy_action)
    problem.add_action(attack_until_conquering_action)
    problem.add_action(conquer_action)

    # add goal: francesco deploys all his troops
    # problem.add_goal(ups.Equals(DeployableTroops(francesco), 0))
    problem.add_goal(Owns(francesco, regions["ontario"]))

    return players


def main():
    problem = ups.Problem("Risk")

    regions = init_map(os.path.join(FILE_DIR, "..", "..", "map.json"), problem)
    players = init_game(problem, regions)
    # print(problem)

    # write pddl file
    # upi.PDDLWriter(problem).write_domain(os.path.join(FILE_DIR, "risk_domain.pddl"))
    # upi.PDDLWriter(problem).write_problem(os.path.join(FILE_DIR, "risk_problem.pddl"))

    with ups.OneshotPlanner(name="enhsp") as planner:
        result = planner.solve(problem)
        if result.status == upe.PlanGenerationResultStatus.SOLVED_SATISFICING:
            print("Planner returned: %s" % result.plan)
        else:
            print("No plan found: {} {}".format(result.status, result.log_messages))

    # plan = upp.SequentialPlan([
    #     deploy_action(players[0], regions["alaska"]),
    #     attack_until_conquering_action(players[0], regions["greenland"], regions["ontario"]),
    #     conquer_action(players[0], players[1], regions["greenland"], regions["ontario"]),
    # ])
    #
    # with ups.SequentialSimulator(problem) as simulator:
    #     simulator: upe.sequential_simulator.UPSequentialSimulator
    #     state = simulator.get_initial_state()
    #     print(f"Initial state: {state}")
    #
    #     for action in plan.actions:
    #         print(f"Applying action: {action}")
    #         state = simulator.apply(state, action)
    #         # print(f"New state: {state}")
    #         # print(f"Remaining battery: {state.get_value(battery)}")
    #     if simulator.is_goal(state):
    #         print("Goal reached!")
    #     else:
    #         print("Goal not reached: {}".format(simulator.get_unsatisfied_goals(state)))


if __name__ == "__main__":
    main()
