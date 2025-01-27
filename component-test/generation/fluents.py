from unified_planning import shortcuts as ups

from user_types import Region, Player, Continent, Phase

# Region related fluents
Adjacent = ups.Fluent("Adjacent", ups.BoolType(), region1=Region, region2=Region)
DeployableTroops = ups.Fluent("DeployableTroops", ups.IntType(), player=Player)
Owns = ups.Fluent("Owns", ups.BoolType(), player=Player, region=Region)
TroopsOn = ups.Fluent("TroopsOn", ups.IntType(), region=Region)

# Continent related fluents
BelongsTo = ups.Fluent("BelongsTo", ups.BoolType(), region=Region, continent=Continent)
BonusTroops = ups.Fluent("BonusTroops", ups.IntType(), continent=Continent)

# Phase/turn related fluents
Turn = ups.Fluent("Turn", ups.BoolType(), player=Player)
NextPlayer = ups.Fluent("NextPlayer", ups.BoolType(), current=Player, next=Player)
CurrentPhase = ups.Fluent("CurrentPhase", ups.BoolType(), phase=Phase)

# Conquer-related fluents
HasWonAttack = ups.Fluent("hasWonAttack", ups.BoolType(), from_region=Region, to_region=Region)
