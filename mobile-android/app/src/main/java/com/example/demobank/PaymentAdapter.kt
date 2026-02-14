package com.example.demobank

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView

class PaymentAdapter(private val payments: List<Payment>) : RecyclerView.Adapter<PaymentAdapter.PaymentViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): PaymentViewHolder {
        val view = LayoutInflater.from(parent.context).inflate(R.layout.item_payment, parent, false)
        return PaymentViewHolder(view)
    }

    override fun onBindViewHolder(holder: PaymentViewHolder, position: Int) {
        val payment = payments[position]
        holder.paymentAmount.text = "$${String.format("%,.2f", payment.amount)}"
        holder.paymentRecipient.text = payment.recipient
        holder.paymentDate.text = payment.date
    }

    override fun getItemCount() = payments.size

    class PaymentViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val paymentAmount: TextView = itemView.findViewById(R.id.payment_amount)
        val paymentRecipient: TextView = itemView.findViewById(R.id.payment_recipient)
        val paymentDate: TextView = itemView.findViewById(R.id.payment_date)
    }
}
