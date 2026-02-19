package com.example.demobank

import android.content.Intent
import android.os.Bundle
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.google.android.material.floatingactionbutton.FloatingActionButton
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class PaymentsFragment : Fragment() {

    private lateinit var recyclerView: RecyclerView
    private var token: String? = null

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
        val view = inflater.inflate(R.layout.fragment_payments, container, false)
        recyclerView = view.findViewById(R.id.payments_recycler_view)
        recyclerView.layoutManager = LinearLayoutManager(context)

        val newPaymentFab = view.findViewById<FloatingActionButton>(R.id.new_payment_fab)
        newPaymentFab.setOnClickListener {
            val intent = Intent(activity, NewPaymentActivity::class.java)
            startActivity(intent)
        }

        return view
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        if (token != null) {
            fetchPayments()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }
    }

    private fun fetchPayments() {
        token?.let {
            DataRepository.getPayments(it).enqueue(object : Callback<PaymentResponse> {
                override fun onResponse(call: Call<PaymentResponse>, response: Response<PaymentResponse>) {
                    if (response.isSuccessful) {
                        val paymentResponse = response.body()
                        if (paymentResponse != null) {
                            recyclerView.adapter = PaymentAdapter(paymentResponse.payments)
                        } else {
                            Toast.makeText(context, "No payments found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch payments: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<PaymentResponse>, t: Throwable) {
                    Log.e("PaymentsFragment", "Failed to fetch payments", t)
                    Toast.makeText(context, "Failed to fetch payments: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
