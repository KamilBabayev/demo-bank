package com.example.demobank

import com.google.gson.GsonBuilder
import com.google.gson.reflect.TypeToken
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object RetrofitClient {

    private const val BASE_URL = "http://10.0.2.2:8080/"

    val instance: ApiService by lazy {
        val gson = GsonBuilder()
            .registerTypeAdapter(object : TypeToken<List<Account>>() {}.type, AccountListDeserializer())
            .registerTypeAdapter(object : TypeToken<List<Card>>() {}.type, CardListDeserializer())
            .create()

        val retrofit = Retrofit.Builder()
            .baseUrl(BASE_URL)
            .addConverterFactory(GsonConverterFactory.create(gson))
            .build()

        retrofit.create(ApiService::class.java)
    }
}
