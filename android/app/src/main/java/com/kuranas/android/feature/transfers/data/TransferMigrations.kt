package com.kuranas.android.feature.transfers.data

import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase

object TransferMigrations {

    val FROM_1_TO_2 = object : Migration(1, 2) {
        override fun migrate(db: SupportSQLiteDatabase) {
            db.execSQL("ALTER TABLE transfers ADD COLUMN mimeType TEXT")
        }
    }

    val FROM_2_TO_3 = object : Migration(2, 3) {
        override fun migrate(db: SupportSQLiteDatabase) {
            db.execSQL("ALTER TABLE transfers ADD COLUMN relativePath TEXT")
        }
    }

    val FROM_3_TO_4 = object : Migration(3, 4) {
        override fun migrate(db: SupportSQLiteDatabase) {
            db.execSQL("ALTER TABLE transfers ADD COLUMN requiresUnmeteredNetwork INTEGER NOT NULL DEFAULT 0")
            db.execSQL(
                "CREATE TABLE IF NOT EXISTS `camera_uploads` (`mediaStoreId` INTEGER NOT NULL, `mediaKind` TEXT NOT NULL, " +
                    "`transferId` INTEGER NOT NULL, `displayName` TEXT NOT NULL, `dateTakenMillis` INTEGER NOT NULL, " +
                    "`registeredAt` INTEGER NOT NULL, PRIMARY KEY(`mediaStoreId`, `mediaKind`))",
            )
            db.execSQL("CREATE INDEX IF NOT EXISTS `index_camera_uploads_transferId` ON `camera_uploads` (`transferId`)")
        }
    }
}
