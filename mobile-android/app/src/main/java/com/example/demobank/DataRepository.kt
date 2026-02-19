package com.example.demobank

import retrofit2.Call

object DataRepository {

    private val apiService: ApiService by lazy { RetrofitClient.instance }

    fun getAccounts(token: String): Call<AccountResponse> {
        return apiService.getAccounts("Bearer $token")
    }

    fun getCards(token: String): Call<CardResponse> {
        return apiService.getCards("Bearer $token")
    }

    fun getPayments(token: String): Call<PaymentResponse> {
        return apiService.getPayments("Bearer $token")
    }

    fun getTransfers(token: String): Call<TransferResponse> {
        return apiService.getTransfers("Bearer $token")
    }

    fun getNotifications(token: String): Call<NotificationResponse> {
        return apiService.getNotifications("Bearer $token")
    }

    fun getMobileOperators(token: String): Call<List<MobileOperator>> {
        return apiService.getMobileOperators("Bearer $token")
    }
}
