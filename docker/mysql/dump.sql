CREATE TABLE `venue` (
  `table_number` INT NOT NULL auto_increment,
  `seats` INT NOT NULL DEFAULT 6,

  PRIMARY KEY (`table_number`)
);

CREATE TABLE `guestlist` (
  `id` INT NOT NULL auto_increment,
  `guest_name` VARCHAR (64) CHARACTER SET utf8 UNIQUE,
  `table_number` INT NOT NULL,
  `accompanying_guests` INT NOT NULL, 
  `time_arrived` TIMESTAMP,
  `arrived` BOOLEAN DEFAULT FALSE,
  
  PRIMARY KEY (`id`),
  FOREIGN KEY (`table_number`) REFERENCES `venue`(`table_number`)
);


/* Unnecessary complexity 
CREATE TABLE `guests` (
  `id` INT NOT NULL auto_increment,
  `name` TEXT,
  `accompanying_guests` INT NOT NULL, 
  PRIMARY KEY (`id`)
);
*/