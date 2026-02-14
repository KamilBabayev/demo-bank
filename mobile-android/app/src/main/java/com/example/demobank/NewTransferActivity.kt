package com.example.demobank

import android.app.Activity
import android.content.Context
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.view.Menu
import android.view.MenuItem
import android.view.View
import android.widget.ArrayAdapter
import android.widget.Button
import android.widget.ProgressBar
import android.widget.Spinner
import android.widget.Toast
import androidx.appcompat.widget.Toolbar
import com.google.android.material.textfield.TextInputEditText
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class NewTransferActivity : AppCompatActivity() {

    private val apiService: ApiService by lazy { RetrofitClient.instance }
    private var token: String? = null
    private var accounts: List<Account> = emptyList()
    private lateinit var progressBar: ProgressBar

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_new_transfer)

        val toolbar = findViewById<Toolbar>(R.id.toolbar)
        setSupportActionBar(toolbar)
        supportActionBar?.setDisplayHomeAsUpEnabled(true)

        val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
        token = sharedPref.getString("TOKEN", null)

        val fromAccountSpinner = findViewById<Spinner>(R.id.from_account_spinner)
        val toAccountEditText = findViewById<TextInputEditText>(R.id.to_account)
        val amountEditText = findViewById<TextInputEditText>(R.id.amount)
        val submitTransferButton = findViewById<Button>(R.id.submit_transfer_button)
        progressBar = findViewById(R.id.progress_bar) // You will need to add this to your layout

        submitTransferButton.setOnClickListener {
            val selectedAccountPosition = fromAccountSpinner.selectedItemPosition
            if (selectedAccountPosition >= 0 && selectedAccountPosition < accounts.size) {
                val fromAccount = accounts[selectedAccountPosition]
                val toAccountId = toAccountEditText.text.toString().toLongOrNull()
                val amount = amountEditText.text.toString().toDoubleOrNull()

                if (toAccountId != null && amount != null) {
                    val newTransferRequest = NewTransferRequest(fromAccount.id, toAccountId, amount)
                    createTransfer(newTransferRequest)
                } else {
                    Toast.makeText(this, "Please enter valid information", Toast.LENGTH_SHORT).show()
                }
            } else {
                Toast.makeText(this, "Please select an account", Toast.LENGTH_SHORT).show()
            }
        }

        fetchAccounts(fromAccountSpinner)
    }

    private fun fetchAccounts(spinner: Spinner) {
        token?.let {
            DataRepository.getAccounts(it).enqueue(object : Callback<AccountResponse> {
                override fun onResponse(call: Call<AccountResponse>, response: Response<AccountResponse>) {
                    if (response.isSuccessful) {
                        val accountResponse = response.body()
                        if (accountResponse != null) {
                            accounts = accountResponse.accounts
                            val adapter = ArrayAdapter(this@NewTransferActivity, R.layout.item_account_spinner, accounts.map { it.account_type })
                            spinner.adapter = adapter
                        } else {
                            Toast.makeText(this@NewTransferActivity, "No accounts found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(this@NewTransferActivity, "Failed to fetch accounts: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<AccountResponse>, t: Throwable) {
                    Toast.makeText(this@NewTransferActivity, "Failed to fetch accounts: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    private fun createTransfer(newTransferRequest: NewTransferRequest) {
        progressBar.visibility = View.VISIBLE
        token?.let {
            apiService.createTransfer("Bearer $it", newTransferRequest).enqueue(object : Callback<Void> {
                override fun onResponse(call: Call<Void>, response: Response<Void>) {
                    progressBar.visibility = View.GONE
                    if (response.isSuccessful) {
                        Toast.makeText(this@NewTransferActivity, "Transfer successful!", Toast.LENGTH_SHORT).show()
                        setResult(Activity.RESULT_OK)
                        finish()
                    } else {
                        Toast.makeText(this@NewTransferActivity, "Transfer failed: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<Void>, t: Throwable) {
                    progressBar.visibility = View.GONE
                    Toast.makeText(this@NewTransferActivity, "Transfer failed: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    override fun onCreateOptionsMenu(menu: Menu?): Boolean {
        menuInflater.inflate(R.menu.new_transfer_menu, menu)
        return true
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        if (item.itemId == R.id.action_history) {
            // Handle history button click
            Toast.makeText(this, "History clicked", Toast.LENGTH_SHORT).show()
            return true
        }
        return super.onOptionsItemSelected(item)
    }

    override fun onSupportNavigateUp(): Boolean {
        onBackPressed()
        return true
    }
}
