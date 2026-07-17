alter table `extendedgames` add column `RematchGameID` varchar(16) not null default '';

alter table `extendedgames` add column `RematchReady` boolean not null default false;

create index idx_extendedgames_rematch_game_id on `extendedgames` (`RematchGameID`);
