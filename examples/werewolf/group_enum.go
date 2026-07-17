package werewolf

// Combining role into the conventional group enum lets state and move
// sanitization use computed relationships such as same-role. In this game a
// shared werewolf role is also the private team relationship.
var groupEnum = enums.MustCombine("group", roleEnum)
