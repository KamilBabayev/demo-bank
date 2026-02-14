package com.example.demobank

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.google.android.material.floatingactionbutton.FloatingActionButton
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class TransfersFragment : Fragment() {

    private lateinit var recyclerView: RecyclerView
    private var token: String? = null
    private var accounts: List<Account> = emptyList()

    private val newTransferLauncher = registerForActivityResult(ActivityResultContracts.StartActivityForResult()) {
        if (it.resultCode == Activity.RESULT_OK) {
            fetchAccounts() // Re-fetch accounts and then transfers to get the latest data
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            token = it.getString("TOKEN")
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        val view = inflater.inflate(R.layout.fragment_transfers, container, false)
        recyclerView = view.findViewById(R.id.transfers_recycler_view)
        recyclerView.layoutManager = LinearLayoutManager(context)

        val newTransferFab = view.findViewById<FloatingActionButton>(R.id.new_transfer_fab)
        newTransferFab.setOnClickListener {
            val intent = Intent(activity, NewTransferActivity::class.java)
            newTransferLauncher.launch(intent)
        }

        return view
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        if (token != null) {
            fetchAccounts()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }
    }

    private fun fetchAccounts() {
        token?.let {
            DataRepository.getAccounts(it).enqueue(object : Callback<AccountResponse> {
                override fun onResponse(call: Call<AccountResponse>, response: Response<AccountResponse>) {
                    if (response.isSuccessful) {
                        val accountResponse = response.body()
                        if (accountResponse != null) {
                            accounts = accountResponse.accounts
                            fetchTransfers()
                        } else {
                            Toast.makeText(context, "No accounts found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch accounts: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<AccountResponse>, t: Throwable) {
                    Toast.makeText(context, "Failed to fetch accounts: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    private fun fetchTransfers() {
        token?.let {
            DataRepository.getTransfers(it).enqueue(object : Callback<TransferResponse> {
                override fun onResponse(call: Call<TransferResponse>, response: Response<TransferResponse>) {
                    if (response.isSuccessful) {
                        val transferResponse = response.body()
                        if (transferResponse != null) {
                            recyclerView.adapter = TransferAdapter(transferResponse.transfers, accounts)
                        } else {
                            Toast.makeText(context, "No transfers found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch transfers: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<TransferResponse>, t: Throwable) {
                    Log.e("TransfersFragment", "Failed to fetch transfers", t)
                    Toast.makeText(context, "Failed to fetch transfers: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
