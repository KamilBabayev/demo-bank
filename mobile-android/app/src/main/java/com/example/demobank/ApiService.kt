package com.example.demobank

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST

@Parcelize
data class LoginRequest(
    val username: String,
    val password: String
) : Parcelable

@Parcelize
data class User(
    val id: Long,
    val username: String,
    val role: String
) : Parcelable

@Parcelize
data class LoginResponse(
    val token: String,
    val expires_at: String,
    val user: User
) : Parcelable

@Parcelize
data class Account(
    val id: Long,
    val user_id: Long,
    val account_number: String,
    val account_type: String,
    val balance: String,
    val currency: String,
    val status: String,
    val created_at: String,
    val updated_at: String
) : Parcelable

@Parcelize
data class Card(
    val id: Long,
    val account_id: Long,
    val card_number: String,
    val card_type: String,
    val cardholder_name: String,
    val expiration_month: Int,
    val expiration_year: Int,
    val status: String,
    val daily_limit: String,
    val monthly_limit: String,
    val per_transaction_limit: String,
    val daily_used: String,
    val monthly_used: String,
    val created_at: String,
    val updated_at: String
) : Parcelable

@Parcelize
data class Payment(
    val id: Long,
    val account_id: Long,
    val amount: Double,
    @SerializedName("recipient_name") val recipient: String,
    @SerializedName("created_at") val date: String
) : Parcelable

@Parcelize
data class Transfer(
    val id: Long,
    val reference_id: String,
    val from_account_id: Long,
    val to_account_id: Long,
    val amount: String,
    val currency: String,
    val status: String,
    val created_at: String,
    val updated_at: String,
    val completed_at: String
) : Parcelable

@Parcelize
data class NewTransferRequest(
    val from_account_id: Long,
    val to_account_id: Long,
    val amount: Double
) : Parcelable

@Parcelize
data class NewPaymentRequest(
    val recipient: String,
    val amount: Double,
    val account_id: Long,
    val payment_type: String,
    val recipient_account: String
) : Parcelable

@Parcelize
data class Notification(
    val id: Long,
    val user_id: Long,
    val type: String,
    val channel: String,
    val title: String,
    val content: String,
    val metadata: NotificationMetadata,
    val status: String,
    @SerializedName("created_at") val createdAt: String
) : Parcelable

@Parcelize
data class NotificationMetadata(
    val payment_id: Long?,
    val transfer_id: Long?,
    val reference_id: String?,
    val failure_reason: String?
) : Parcelable

@Parcelize
data class MobileOperator(
    val id: Long,
    val name: String,
    val prefixes: List<String>
) : Parcelable

interface ApiService {
    @POST("/api/v1/auth/login")
    fun login(@Body request: LoginRequest): Call<LoginResponse>

    @GET("/api/v1/accounts")
    fun getAccounts(@Header("Authorization") token: String): Call<AccountResponse>

    @GET("/api/v1/cards")
    fun getCards(@Header("Authorization") token: String): Call<CardResponse>

    @GET("/api/v1/payments")
    fun getPayments(@Header("Authorization") token: String): Call<PaymentResponse>

    @GET("/api/v1/transfers")
    fun getTransfers(@Header("Authorization") token: String): Call<TransferResponse>

    @POST("/api/v1/transfers")
    fun createTransfer(@Header("Authorization") token: String, @Body request: NewTransferRequest): Call<Void>

    @GET("/api/v1/notifications")
    fun getNotifications(@Header("Authorization") token: String): Call<NotificationResponse>

    @POST("/api/v1/payments")
    fun createPayment(@Header("Authorization") token: String, @Body request: NewPaymentRequest): Call<Void>

    @GET("/api/v1/mobile-operators")
    fun getMobileOperators(@Header("Authorization") token: String): Call<List<MobileOperator>>
}
