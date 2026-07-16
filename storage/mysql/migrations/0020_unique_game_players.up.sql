-- Older servers allowed both duplicate users and overwritten/duplicate seat
-- rows. Retain the earliest association deterministically before installing
-- the invariants; otherwise the migration would fail and leave deployments on
-- the unsafe schema.
delete newer from `players` newer
join `players` older
  on newer.`GameID` = older.`GameID`
 and newer.`UserID` = older.`UserID`
 and newer.`Id` > older.`Id`;

delete newer from `players` newer
join `players` older
  on newer.`GameID` = older.`GameID`
 and newer.`PlayerIndex` = older.`PlayerIndex`
 and newer.`Id` > older.`Id`;

alter table `players`
  add unique key `uniq_players_game_player` (`GameID`, `PlayerIndex`),
  add unique key `uniq_players_game_user` (`GameID`, `UserID`);
