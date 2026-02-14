package com.example.demobank

import android.os.Bundle
import android.util.Log
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

class DashboardFragment : Fragment() {

    private lateinit var welcomeMessage: TextView
    private lateinit var totalBalance: TextView
    private lateinit var accountsRecyclerView: RecyclerView
    private var token: String? = null
    private var username: String? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            token = it.getString("TOKEN")
            username = it.getString("USERNAME")
        }
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

        welcomeMessage.text = "Welcome, $username"

        return view
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        if (token != null) {
            fetchDashboardData()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }
    }

    private fun fetchDashboardData() {
        token?.let {
            DataRepository.getAccounts(it).enqueue(object : Callback<AccountResponse> {
                override fun onResponse(call: Call<AccountResponse>, response: Response<AccountResponse>) {
                    if (response.isSuccessful) {
                        val accountResponse = response.body()
                        if (accountResponse != null) {
                            accountsRecyclerView.adapter = AccountAdapter(accountResponse.accounts)
                            val total = accountResponse.accounts.sumOf { it.balance.toDouble() }
                            totalBalance.text = "$${String.format("%,.2f", total)}"
                        } else {
                            Toast.makeText(context, "No accounts found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch accounts: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<AccountResponse>, t: Throwable) {
                    Log.e("DashboardFragment", "Failed to fetch accounts", t)
                    Toast.makeText(context, "Failed to fetch accounts: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
