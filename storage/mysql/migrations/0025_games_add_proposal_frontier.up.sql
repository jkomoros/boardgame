alter table `games` add column `ProposalFrontierVersion` bigint not null default 0;
alter table `games` add column `ProposalFrontierKnown` boolean not null default false;
