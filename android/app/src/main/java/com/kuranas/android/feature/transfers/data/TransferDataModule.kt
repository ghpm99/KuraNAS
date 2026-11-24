package com.kuranas.android.feature.transfers.data

import android.content.Context
import androidx.room.Room
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object TransferDataModule {

    @Provides
    @Singleton
    fun provideKuranasDatabase(@ApplicationContext context: Context): KuranasDatabase =
        Room.databaseBuilder(context, KuranasDatabase::class.java, "kuranas.db")
            .addMigrations(TransferMigrations.FROM_1_TO_2, TransferMigrations.FROM_2_TO_3)
            .build()

    @Provides
    @Singleton
    fun provideTransferDao(database: KuranasDatabase): TransferDao = database.transferDao()

    @Provides
    fun provideTransferClock(): TransferClock = TransferClock { System.currentTimeMillis() }
}
