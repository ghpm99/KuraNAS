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
}
