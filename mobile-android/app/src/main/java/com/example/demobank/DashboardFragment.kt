package com.example.demobank

import android.os.Bundle
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import android.widget.Toast
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

class DashboardFragment : Fragment() {

    private lateinit var welcomeMessage: TextView
    private lateinit var totalBalance: TextView
    private lateinit var accountsRecyclerView: RecyclerView
    private lateinit var apiService: ApiService
    private var token: String? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            token = it.getString("TOKEN")
        }

        val retrofit = Retrofit.Builder()
            .baseUrl("http://10.0.2.2:8080") // Use 10.0.2.2 for localhost from Android emulator
            .addConverterFactory(GsonConverterFactory.create())
            .build()

        apiService = retrofit.create(ApiService::class.java)
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        val view = inflater.inflate(R.layout.fragment_dashboard, container, false)
        welcomeMessage = view.findViewById(R.id.welcome_message)
        totalBalance = view.findViewById(R.id.total_balance)
        accountsRecyclerView = view.findViewById(R.id.accounts_recycler_view)
        accountsRecyclerView.layoutManager = LinearLayoutManager(context)

        if (token != null) {
            fetchDashboardData()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }

        return view
    }

    private fun fetchDashboardData() {
        token?.let {
            apiService.getAccounts("Bearer $it").enqueue(object : Callback<List<Account>> {
                override fun onResponse(call: Call<List<Account>>, response: Response<List<Account>>) {
                    if (response.isSuccessful) {
                        val accounts = response.body()
                        if (accounts != null) {
                            accountsRecyclerView.adapter = AccountAdapter(accounts)
                            val total = accounts.sumOf { it.balance }
                            totalBalance.text = "$${String.format("%,.2f", total)}"
                        } else {
                            Toast.makeText(context, "No accounts found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch accounts: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<List<Account>>, t: Throwable) {
                    Toast.makeText(context, "Failed to fetch accounts: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
