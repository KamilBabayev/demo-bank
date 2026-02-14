package com.example.demobank

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView

data class Card(
    val card_number: String,
    val expiration_date: String,
    val cardholder_name: String
)

class CardAdapter(private val cards: List<Card>) : RecyclerView.Adapter<CardAdapter.CardViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): CardViewHolder {
        val view = LayoutInflater.from(parent.context).inflate(R.layout.item_card, parent, false)
        return CardViewHolder(view)
    }

    override fun onBindViewHolder(holder: CardViewHolder, position: Int) {
        val card = cards[position]
        holder.cardNumber.text = card.card_number
        holder.cardExpiration.text = card.expiration_date
        holder.cardHolder.text = card.cardholder_name
    }

    override fun getItemCount() = cards.size

    class CardViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val cardNumber: TextView = itemView.findViewById(R.id.card_number)
        val cardExpiration: TextView = itemView.findViewById(R.id.card_expiration)
        val cardHolder: TextView = itemView.findViewById(R.id.card_holder)
    }
}
