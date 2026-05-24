alter table `extendedgames` add column `CompanionRoomCode` varchar(8) not null default '';
alter table `extendedgames` add column `CompanionLocked` boolean not null default false;
-- Unique index on non-empty room codes only would be ideal but MySQL doesn't
-- support partial indexes; a regular index supports the GameByRoomCode lookup.
-- Collisions are prevented at write time by retry logic in the room-code
-- generator (see server/api/roomcode.go).
create index idx_extendedgames_companion_room_code on `extendedgames` (`CompanionRoomCode`);
