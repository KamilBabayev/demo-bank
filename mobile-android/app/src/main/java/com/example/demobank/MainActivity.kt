package com.example.demobank

import android.content.Context
import android.content.Intent
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.widget.Button
import android.widget.Toast
import com.google.android.material.textfield.TextInputEditText
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class MainActivity : AppCompatActivity() {

    private val apiService: ApiService by lazy { RetrofitClient.instance }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val usernameEditText = findViewById<TextInputEditText>(R.id.username)
        val passwordEditText = findViewById<TextInputEditText>(R.id.password)
        val loginButton = findViewById<Button>(R.id.loginButton)

        loginButton.setOnClickListener {
            val username = usernameEditText.text.toString()
            val password = passwordEditText.text.toString()

            if (username.isNotEmpty() && password.isNotEmpty()) {
                val loginRequest = LoginRequest(username, password)
                login(loginRequest)
            } else {
                Toast.makeText(this, "Please enter username and password", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun login(loginRequest: LoginRequest) {
        apiService.login(loginRequest).enqueue(object : Callback<LoginResponse> {
            override fun onResponse(call: Call<LoginResponse>, response: Response<LoginResponse>) {
                if (response.isSuccessful) {
                    val loginResponse = response.body()
                    if (loginResponse != null) {
                        // Login successful, save token and start HomeActivity
                        val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
                        with(sharedPref.edit()) {
                            putString("TOKEN", loginResponse.token)
                            apply()
                        }

                        val intent = Intent(this@MainActivity, HomeActivity::class.java)
                        intent.putExtra("USERNAME", loginResponse.user.username)
                        startActivity(intent)
                        finish() // Finish MainActivity so the user can't go back to the login screen
                    } else {
                        Toast.makeText(this@MainActivity, "Login failed: Empty response", Toast.LENGTH_SHORT).show()
                    }
                } else {
                    // Login failed
                    Toast.makeText(this@MainActivity, "Login failed: " + response.message(), Toast.LENGTH_SHORT).show()
                }
            }

            override fun onFailure(call: Call<LoginResponse>, t: Throwable) {
                // Network error or other failure
                Toast.makeText(this@MainActivity, "Login failed: " + t.message, Toast.LENGTH_SHORT).show()
            }
        })
    }
}
