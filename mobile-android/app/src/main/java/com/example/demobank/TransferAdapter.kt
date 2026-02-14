package com.example.demobank

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView

class TransferAdapter(private val transfers: List<Transfer>) : RecyclerView.Adapter<TransferAdapter.TransferViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): TransferViewHolder {
        val view = LayoutInflater.from(parent.context).inflate(R.layout.item_transfer, parent, false)
        return TransferViewHolder(view)
    }

    override fun onBindViewHolder(holder: TransferViewHolder, position: Int) {
        val transfer = transfers[position]
        holder.transferAmount.text = "$${String.format("%,.2f", transfer.amount.toDouble())}"
        holder.transferFrom.text = "From: ${transfer.from_account_id}"
        holder.transferTo.text = "To: ${transfer.to_account_id}"
        holder.transferDate.text = transfer.date
    }

    override fun getItemCount() = transfers.size

    class TransferViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val transferAmount: TextView = itemView.findViewById(R.id.transfer_amount)
        val transferFrom: TextView = itemView.findViewById(R.id.transfer_from)
        val transferTo: TextView = itemView.findViewById(R.id.transfer_to)
        val transferDate: TextView = itemView.findViewById(R.id.transfer_date)
    }
}
