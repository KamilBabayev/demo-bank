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

class NotificationsFragment : Fragment() {

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
        val view = inflater.inflate(R.layout.fragment_notifications, container, false)
        recyclerView = view.findViewById(R.id.notifications_recycler_view)
        recyclerView.layoutManager = LinearLayoutManager(context)

        return view
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        if (token != null) {
            fetchNotifications()
        } else {
            Toast.makeText(context, "Authentication token not found", Toast.LENGTH_SHORT).show()
        }
    }

    private fun fetchNotifications() {
        token?.let {
            DataRepository.getNotifications(it).enqueue(object : Callback<NotificationResponse> {
                override fun onResponse(call: Call<NotificationResponse>, response: Response<NotificationResponse>) {
                    if (response.isSuccessful) {
                        val notificationResponse = response.body()
                        if (notificationResponse != null) {
                            recyclerView.adapter = NotificationAdapter(notificationResponse.notifications)
                        } else {
                            Toast.makeText(context, "No notifications found", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(context, "Failed to fetch notifications: " + response.message(), Toast.LENGTH_SHORT).show()
                    }
                }

                override fun onFailure(call: Call<NotificationResponse>, t: Throwable) {
                    Log.e("NotificationsFragment", "Failed to fetch notifications", t)
                    Toast.makeText(context, "Failed to fetch notifications: " + t.message, Toast.LENGTH_SHORT).show()
                }
            })
        }
    }
}
