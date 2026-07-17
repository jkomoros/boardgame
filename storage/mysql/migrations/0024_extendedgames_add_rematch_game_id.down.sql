drop index idx_extendedgames_rematch_game_id on `extendedgames`;

alter table `extendedgames` drop column `RematchReady`;

alter table `extendedgames` drop column `RematchGameID`;
