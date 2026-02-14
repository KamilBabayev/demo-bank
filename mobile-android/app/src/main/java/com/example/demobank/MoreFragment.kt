package com.example.demobank

import android.content.Context
import android.content.Intent
import android.os.Bundle
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView

class MoreFragment : Fragment() {

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        val view = inflater.inflate(R.layout.fragment_more, container, false)

        val notificationsButton = view.findViewById<TextView>(R.id.notifications_button)
        notificationsButton.setOnClickListener {
            val fragment = NotificationsFragment()
            fragment.arguments = arguments
            parentFragmentManager.beginTransaction().replace(R.id.fragment_container, fragment).commit()
        }

        val logoutButton = view.findViewById<TextView>(R.id.logout_button)
        logoutButton.setOnClickListener {
            val sharedPref = activity?.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
            val editor = sharedPref?.edit()
            editor?.remove("TOKEN")
            editor?.apply()

            val intent = Intent(activity, MainActivity::class.java)
            startActivity(intent)
            activity?.finish()
        }

        return view
    }
}
