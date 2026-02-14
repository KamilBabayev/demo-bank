package com.example.demobank

import com.google.gson.JsonElement
import retrofit2.Call

object DataRepository {

    private val apiService: ApiService by lazy { RetrofitClient.instance }

    fun getAccounts(token: String): Call<JsonElement> {
        return apiService.getAccounts("Bearer $token")
    }

    fun getCards(token: String): Call<JsonElement> {
        return apiService.getCards("Bearer $token")
    }

    fun getPayments(token: String): Call<JsonElement> {
        return apiService.getPayments("Bearer $token")
    }
}
