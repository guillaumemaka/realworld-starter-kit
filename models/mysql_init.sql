CREATE DATABASE IF NOT EXISTS conduit;

CREATE TABLE IF NOT EXISTS `articles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `slug` varchar(255) NOT NULL DEFAULT '',
  `title` varchar(255) NOT NULL DEFAULT '',
  `description` varchar(255) NOT NULL DEFAULT '',
  `body` text NOT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `author_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `slug` (`slug`)
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
  PRIMARY KEY (`usr_id`,`art_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `usr_following` (
  `usr_id` int(11) unsigned NOT NULL,
  `usr_following_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`usr_id`,`usr_following_id`),
  KEY `following_id` (`usr_following_id`),
  CONSTRAINT `following_id` FOREIGN KEY (`usr_following_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_id` FOREIGN KEY (`usr_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;