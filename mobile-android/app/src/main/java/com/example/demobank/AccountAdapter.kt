package com.example.demobank

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView

class AccountAdapter(private val accounts: List<Account>) : RecyclerView.Adapter<AccountAdapter.AccountViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): AccountViewHolder {
        val view = LayoutInflater.from(parent.context).inflate(R.layout.item_account, parent, false)
        return AccountViewHolder(view)
    }

    override fun onBindViewHolder(holder: AccountViewHolder, position: Int) {
        val account = accounts[position]
        holder.accountName.text = account.account_type
        holder.accountBalance.text = "${account.balance} ${account.currency}"
    }

    override fun getItemCount() = accounts.size

    class AccountViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val accountName: TextView = itemView.findViewById(R.id.account_name)
        val accountBalance: TextView = itemView.findViewById(R.id.account_balance)
    }
}
