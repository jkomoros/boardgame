drop index idx_extendedgames_companion_room_code on `extendedgames`;
alter table `extendedgames` drop column `CompanionLocked`;
alter table `extendedgames` drop column `CompanionRoomCode`;
