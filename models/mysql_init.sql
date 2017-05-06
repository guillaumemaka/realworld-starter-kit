CREATE DATABASE IF NOT EXISTS conduit CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE IF NOT EXISTS `articles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `slug` varchar(255) NOT NULL DEFAULT '',
  `title` varchar(255) NOT NULL DEFAULT '',
  `description` varchar(255) NOT NULL DEFAULT '',
  `body` text NOT NULL,
  `created` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `author_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `slug` (`slug`),
  KEY `ARTICLEusrID` (`author_id`),
  CONSTRAINT `ARTICLEusrID` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL DEFAULT '',
  `email` varchar(255) NOT NULL DEFAULT '',
  `password` char(60) NOT NULL DEFAULT '',
  `bio` text NOT NULL,
  `image` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `usr_art_favourite` (
  `usr_id` int(11) unsigned NOT NULL,
  `art_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`usr_id`,`art_id`),
  KEY `FAVOURITEartID` (`art_id`),
  CONSTRAINT `FAVOURITEartID` FOREIGN KEY (`art_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FAVOURITEusrID` FOREIGN KEY (`usr_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `usr_following` (
  `usr_id` int(11) unsigned NOT NULL,
  `usr_following_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`usr_id`,`usr_following_id`),
  KEY `following_id` (`usr_following_id`),
  CONSTRAINT `FOLLOWuserID` FOREIGN KEY (`usr_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FOLLOWusrFollowID` FOREIGN KEY (`usr_following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;