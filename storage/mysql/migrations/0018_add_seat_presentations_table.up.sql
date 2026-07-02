create table if not exists `seatpresentations` (
  `ID` bigint not null primary key auto_increment,
  `GameID` varchar(16) not null,
  `PlayerIndex` bigint not null,
  `DisplayName` varchar(128) not null default '',
  `AvatarSlug` varchar(256) not null default '',
  unique key uniq_seatpresentations_game_player (`GameID`, `PlayerIndex`)
) engine=InnoDB charset=utf8mb4;
