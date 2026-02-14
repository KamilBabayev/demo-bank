package com.example.demobank

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import java.text.SimpleDateFormat
import java.util.Locale

class TransferAdapter(private val transfers: List<Transfer>, private val accounts: List<Account>) : RecyclerView.Adapter<TransferAdapter.TransferViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): TransferViewHolder {
        val view = LayoutInflater.from(parent.context).inflate(R.layout.item_transfer, parent, false)
        return TransferViewHolder(view)
    }

    override fun onBindViewHolder(holder: TransferViewHolder, position: Int) {
        val transfer = transfers[position]
        holder.transferAmount.text = "$${String.format("%,.2f", transfer.amount.toDouble())}"
        holder.transferFrom.text = "From: ${getAccountType(transfer.from_account_id)}"
        holder.transferTo.text = "To: ${getAccountType(transfer.to_account_id)}"
        holder.transferDate.text = formatDate(transfer.created_at)
    }

    override fun getItemCount() = transfers.size

    private fun getAccountType(accountId: Long): String {
        val account = accounts.find { it.id == accountId }
        return account?.account_type ?: "Unknown Account"
    }

    private fun formatDate(dateString: String): String {
        return try {
            val inputFormat = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSSSSS'Z'", Locale.getDefault())
            val outputFormat = SimpleDateFormat("MMM dd, yyyy", Locale.getDefault())
            val date = inputFormat.parse(dateString)
            outputFormat.format(date)
        } catch (e: Exception) {
            dateString // Return original string if parsing fails
        }
    }

    class TransferViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val transferAmount: TextView = itemView.findViewById(R.id.transfer_amount)
        val transferFrom: TextView = itemView.findViewById(R.id.transfer_from)
        val transferTo: TextView = itemView.findViewById(R.id.transfer_to)
        val transferDate: TextView = itemView.findViewById(R.id.transfer_date)
    }
}
