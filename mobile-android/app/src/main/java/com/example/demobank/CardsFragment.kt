package com.example.demobank

import android.os.Bundle
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class CardsFragment : Fragment() {

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
        val view = inflater.inflate(R.layout.fragment_cards, container, false)
        recyclerView = view.findViewById(R.id.cards_recycler_view)
        recyclerView.layoutManager = LinearLayoutManager(context)

        return view
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        if (token != null) {
            fetchCards()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }
    }

    private fun fetchCards() {
        token?.let {
            DataRepository.getCards(it).enqueue(object : Callback<CardResponse> {
                override fun onResponse(call: Call<CardResponse>, response: Response<CardResponse>) {
                    if (response.isSuccessful) {
                        val cardResponse = response.body()
                        if (cardResponse != null) {
                            recyclerView.adapter = CardAdapter(cardResponse.cards)
                        } else {
                            Toast.makeText(context, "No cards found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch cards: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<CardResponse>, t: Throwable) {
                    Log.e("CardsFragment", "Failed to fetch cards", t)
                    Toast.makeText(context, "Failed to fetch cards: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
