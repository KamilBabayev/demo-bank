package com.example.demobank

import android.content.Context
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.view.View
import android.widget.AdapterView
import android.widget.ArrayAdapter
import android.widget.Button
import android.widget.Spinner
import android.widget.Toast
import androidx.appcompat.widget.Toolbar
import com.google.android.material.textfield.TextInputEditText
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class NewPaymentActivity : AppCompatActivity() {

    private val apiService: ApiService by lazy { RetrofitClient.instance }
    private var token: String? = null

    private var accounts: List<Account> = emptyList()
    private var operators: List<MobileOperator> = emptyList()
    private var selectedAccountId: Long? = null
    private var selectedOperator: MobileOperator? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_new_payment)

        val toolbar = findViewById<Toolbar>(R.id.toolbar)
        setSupportActionBar(toolbar)
        supportActionBar?.setDisplayHomeAsUpEnabled(true)

        val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
        token = sharedPref.getString("TOKEN", null)

        val accountSpinner = findViewById<Spinner>(R.id.account_spinner)
        val providerSpinner = findViewById<Spinner>(R.id.provider_spinner)
        val phoneEditText = findViewById<TextInputEditText>(R.id.phone_number)
        val amountEditText = findViewById<TextInputEditText>(R.id.amount)
        val submitPaymentButton = findViewById<Button>(R.id.submit_payment_button)

        fetchAccounts(accountSpinner)
        fetchMobileOperators(providerSpinner)

        submitPaymentButton.setOnClickListener {
            val phone = phoneEditText.text.toString().trim()
            val amount = amountEditText.text.toString().toDoubleOrNull()

            if (selectedAccountId == null) {
                Toast.makeText(this, "Please select an account", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (selectedOperator == null) {
                Toast.makeText(this, "Please select an operator", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (phone.length != 7 || !phone.all { it.isDigit() }) {
                Toast.makeText(this, "Phone number must be exactly 7 digits", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (amount == null || amount <= 0) {
                Toast.makeText(this, "Please enter a valid amount", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            val prefix = selectedOperator!!.prefixes.first()
            val request = NewPaymentRequest(
                account_id = selectedAccountId!!,
                payment_type = "mobile",
                recipient = selectedOperator!!.name,
                recipient_account = prefix + phone,
                amount = amount
            )
            createPayment(request)
        }
    }

    private fun fetchAccounts(spinner: Spinner) {
        token?.let { t ->
            apiService.getAccounts("Bearer $t").enqueue(object : Callback<AccountResponse> {
                override fun onResponse(call: Call<AccountResponse>, response: Response<AccountResponse>) {
                    if (response.isSuccessful) {
                        accounts = response.body()?.accounts?.filter { it.status == "active" } ?: emptyList()
                        val labels = accounts.map { "${it.account_number} (${it.currency} ${it.balance})" }
                        val adapter = ArrayAdapter(this@NewPaymentActivity, R.layout.item_provider_spinner, labels)
                        spinner.adapter = adapter
                        spinner.onItemSelectedListener = object : AdapterView.OnItemSelectedListener {
                            override fun onItemSelected(parent: AdapterView<*>?, view: View?, position: Int, id: Long) {
                                selectedAccountId = accounts[position].id
                            }
                            override fun onNothingSelected(parent: AdapterView<*>?) {
                                selectedAccountId = null
                            }
                        }
                    }
                }
                override fun onFailure(call: Call<AccountResponse>, t: Throwable) {
                    Toast.makeText(this@NewPaymentActivity, "Failed to load accounts", Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    private fun fetchMobileOperators(spinner: Spinner) {
        token?.let { t ->
            apiService.getMobileOperators("Bearer $t").enqueue(object : Callback<List<MobileOperator>> {
                override fun onResponse(call: Call<List<MobileOperator>>, response: Response<List<MobileOperator>>) {
                    if (response.isSuccessful) {
                        operators = response.body() ?: emptyList()
                        val names = operators.map { it.name }
                        val adapter = ArrayAdapter(this@NewPaymentActivity, R.layout.item_provider_spinner, names)
                        spinner.adapter = adapter
                        spinner.onItemSelectedListener = object : AdapterView.OnItemSelectedListener {
                            override fun onItemSelected(parent: AdapterView<*>?, view: View?, position: Int, id: Long) {
                                selectedOperator = operators[position]
                            }
                            override fun onNothingSelected(parent: AdapterView<*>?) {
                                selectedOperator = null
                            }
                        }
                    }
                }
                override fun onFailure(call: Call<List<MobileOperator>>, t: Throwable) {
                    Toast.makeText(this@NewPaymentActivity, "Failed to load operators", Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    private fun createPayment(request: NewPaymentRequest) {
        token?.let {
            apiService.createPayment("Bearer $it", request).enqueue(object : Callback<Void> {
                override fun onResponse(call: Call<Void>, response: Response<Void>) {
                    if (response.isSuccessful) {
                        Toast.makeText(this@NewPaymentActivity, "Payment successful!", Toast.LENGTH_SHORT).show()
                        finish()
                    } else {
                        Toast.makeText(this@NewPaymentActivity, "Payment failed: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<Void>, t: Throwable) {
                    Toast.makeText(this@NewPaymentActivity, "Payment failed: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }

    override fun onSupportNavigateUp(): Boolean {
        onBackPressed()
        return true
    }
}
