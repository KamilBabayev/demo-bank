package com.example.demobank

import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.POST

data class LoginRequest(
    val username: String,
    val password: String
)

data class User(
    val id: Long,
    val username: String,
    val role: String
)

data class LoginResponse(
    val token: String,
    val expires_at: String,
    val user: User
)

interface ApiService {
    @POST("/api/v1/auth/login")
    fun login(@Body request: LoginRequest): Call<LoginResponse>
}
