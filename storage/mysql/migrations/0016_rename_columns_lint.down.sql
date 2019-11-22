alter table `games` change ID Id varchar(16) not null;
alter table `extendedgames` change ID Id varchar(16) not null;
alter table `states` change ID Id bigint not null auto_increment;
alter table `states` change GameID GameId varchar(16) not null;
alter table `moves` change ID Id bigint not null auto_increment;
alter table `moves` change GameID GameId varchar(16) not null;
alter table `players` change ID Id bigint not null auto_increment;
alter table `players` change GameID GameId varchar(16) not null;
alter table `players` change UserID UserId varchar(128);
alter table `agentstates` change ID Id bigint not null auto_increment;
alter table `agentstates` change GameID GameId varchar(16) not null;
alter table `users` change ID Id varchar(128) not null;
alter table `users` change PhotoURL PhotoUrl text;
alter table `cookies` change UserID UserId varchar(128) not null;