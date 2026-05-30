-keep class com.kuranas.android.** { *; }
-keepclassmembers class * {
    @kotlinx.serialization.SerialName *;
}
-keep @kotlinx.serialization.Serializable class * { *; }
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.**
-keepclasseswithmembers class * {
    @com.squareup.moshi.* <methods>;
}
